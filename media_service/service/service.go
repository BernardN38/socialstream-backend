package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"time"

	rabbitmq_producer "github.com/BernardN38/socialstream-backend/media_service/rabbitmq/producer"
	rpc_client "github.com/BernardN38/socialstream-backend/media_service/rpc/client"
	media_sql "github.com/BernardN38/socialstream-backend/media_service/sql/media"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type MediaService struct {
	minioClient      *minio.Client
	mediaDb          *sql.DB
	mediaQueries     *media_sql.Queries
	rpcClient        *rpc_client.RpcClient
	rabbitmqProducer *rabbitmq_producer.RabbitMQProducer
	config           *MediaServiceConfig
}
type ImageUpload struct {
	MediaData     io.Reader
	UserId        int32
	ContentType   string
	ContentLength int64
}
type MediaUpdate struct {
	MediaId       int32     `json:"mediaId"`
	NewExternalId uuid.UUID `json:"newExternalId"`
	OldExternalId uuid.UUID `json:"oldExternalId"`
}
type MediaServiceConfig struct {
	MinioBucketName string
}

func New(minioClient *minio.Client, mediaDb *sql.DB, rpcClient *rpc_client.RpcClient, rabbitmqProducer *rabbitmq_producer.RabbitMQProducer, config MediaServiceConfig) (*MediaService, error) {
	mediaQueries := media_sql.New(mediaDb)

	err := SetupMinio(minioClient, config.MinioBucketName)
	if err != nil {
		return nil, err
	}
	return &MediaService{
		minioClient:      minioClient,
		mediaDb:          mediaDb,
		mediaQueries:     mediaQueries,
		rpcClient:        rpcClient,
		rabbitmqProducer: rabbitmqProducer,
		config:           &config,
	}, nil
}
func (m *MediaService) GetAllMedia(ctx context.Context) ([]media_sql.Medium, error) {
	media, err := m.mediaQueries.GetAllMedia(ctx)
	if err != nil {
		return nil, err
	}
	return media, nil
}
func (m *MediaService) GetExternalId(ctx context.Context, externalId uuid.UUID) (*minio.Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	respCh := make(chan *minio.Object)
	errCh := make(chan error)
	go func() {
		object, err := m.minioClient.GetObject(ctx, "media-service", externalId.String(), minio.GetObjectOptions{})
		if err != nil {
			errCh <- err
		}
		respCh <- object
	}()
	select {
	case object := <-respCh:
		return object, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (m *MediaService) GetMediaCompressed(ctx context.Context, mediaId int32) (*minio.Object, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	respCh := make(chan *minio.Object)
	errCh := make(chan error)

	go func() {
		externalId, err := m.mediaQueries.GetCompressedExternalId(timeoutCtx, mediaId)
		object, err := m.minioClient.GetObject(ctx, "media-service", externalId.String(), minio.GetObjectOptions{})
		if err != nil {
			errCh <- err
		}
		respCh <- object
	}()
	select {
	case object := <-respCh:
		return object, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}
func (m *MediaService) UploadMedia(ctx context.Context, payload ImageUpload) (*int32, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	tx, err := m.mediaDb.BeginTx(timeoutCtx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txQuries := m.mediaQueries.WithTx(tx)
	newExternalIdFull := uuid.New()
	newExternalIdCompressed := uuid.New()
	mediaId, err := txQuries.CreateMedia(ctx, media_sql.CreateMediaParams{
		ExternalUuidFull:       newExternalIdFull,
		ExternalUuidCompressed: newExternalIdCompressed,
		UserID:                 payload.UserId,
		CompressionStatus:      "started",
		IsActive:               true,
	})
	infoCh := make(chan minio.UploadInfo)
	errCh := make(chan error)

	go func() {
		info, err := m.minioClient.PutObject(timeoutCtx, m.config.MinioBucketName, newExternalIdFull.String(), payload.MediaData, payload.ContentLength, minio.PutObjectOptions{
			ContentType: payload.ContentType,
		})
		if err != nil {
			errCh <- err
			return
		}

		msg := struct {
			MediaId              int32     `json:"mediaId"`
			ExternalIdFull       uuid.UUID `json:"externalIdFull"`
			ExternalIdCompressed uuid.UUID `json:"externalIdCompressed"`
			ContentType          string    `json:"contentType"`
		}{
			MediaId:              mediaId,
			ExternalIdFull:       newExternalIdFull,
			ExternalIdCompressed: newExternalIdCompressed,
			ContentType:          payload.ContentType,
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			errCh <- err
			return
		}
		log.Println("media uploaded publishing message:", msg)
		err = m.rabbitmqProducer.Publish("media_events", "media.uploaded", msgBytes)
		if err != nil {
			errCh <- err
			return
		}
		infoCh <- info
	}()

	select {
	case info := <-infoCh:
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		log.Println(info)
		return &mediaId, nil
	case err := <-errCh:
		tx.Rollback()
		return nil, err
	case <-timeoutCtx.Done():
		tx.Rollback()
		return nil, timeoutCtx.Err()
	}
}
func (m *MediaService) DeleteMedia(ctx context.Context, mediaId int32) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	tx, err := m.mediaDb.BeginTx(timeoutCtx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQuries := m.mediaQueries.WithTx(tx)

	errCh := make(chan error)
	successCh := make(chan bool)
	go func() {
		log.Println(mediaId)
		compressionStatus, err := txQuries.GetCompressionStatusById(timeoutCtx, mediaId)
		if err != nil {
			errCh <- err
			return
		}
		if compressionStatus == "complete" {
			externalIds, err := txQuries.GetExternalIdsById(timeoutCtx, mediaId)
			if err != nil {
				errCh <- err
				return
			}
			log.Println("removing exiternal id:", externalIds.ExternalUuidFull.String())
			err = m.minioClient.RemoveObject(timeoutCtx, "media-service", externalIds.ExternalUuidFull.String(), minio.RemoveObjectOptions{})
			if err != nil {
				errCh <- err
				return
			}
			log.Println("removing exiternal id:", externalIds.ExternalUuidCompressed.String())
			err = m.minioClient.RemoveObject(timeoutCtx, "media-service", externalIds.ExternalUuidCompressed.String(), minio.RemoveObjectOptions{})
			if err != nil {
				errCh <- err
				return
			}
			err = txQuries.DeleteMediaById(timeoutCtx, mediaId)
			if err != nil {
				errCh <- err
				return
			}
			successCh <- true
		} else {
			time.Sleep(time.Second * 1)
			errCh <- errors.New("compression not done")
		}
	}()

	select {
	case <-successCh:
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		tx.Rollback()
		return err
	case <-timeoutCtx.Done():
		tx.Rollback()
		return timeoutCtx.Err()
	}
}
func (m *MediaService) DeleteExternalId(ctx context.Context, externalId uuid.UUID) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	successCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		_, err := m.minioClient.StatObject(timeoutCtx, m.config.MinioBucketName, externalId.String(), minio.GetObjectOptions{})
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			errCh <- err
			return
		}
		err = m.minioClient.RemoveObject(timeoutCtx, m.config.MinioBucketName, externalId.String(), minio.RemoveObjectOptions{})
		if err != nil {
			errCh <- err
			return
		}
		successCh <- true
	}()

	select {
	case <-successCh:
		return nil
	case err := <-errCh:
		log.Println(err)
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

func (m *MediaService) GetUserProfileImage(ctx context.Context, userId int32) (*minio.Object, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	respCh := make(chan *minio.Object)
	errCh := make(chan error)
	mediaId, err := m.rpcClient.GetUserProfileImageIdRpc(userId)
	if err != nil {
		return nil, err
	}
	externalIds, err := m.mediaQueries.GetExternalIdsById(ctx, mediaId)
	if err != nil {
		return nil, err
	}
	var retrieveId uuid.UUID
	if externalIds.CompressionStatus == "complete" {
		retrieveId = externalIds.ExternalUuidCompressed
	} else {
		retrieveId = externalIds.ExternalUuidFull
	}
	go func(context.Context) {
		log.Println("getting mediaId: ", mediaId)
		object, err := m.minioClient.GetObject(ctx, "media-service", retrieveId.String(), minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			errCh <- err
			return
		}
		respCh <- object
	}(ctx)
	select {
	case object := <-respCh:
		return object, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	}
}

func (m *MediaService) UploadUserProfileImage(ctx context.Context, userId int32, image multipart.File, imageHeader *multipart.FileHeader) error {

	timeoutCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()

	tx, err := m.mediaDb.BeginTx(timeoutCtx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQuries := m.mediaQueries.WithTx(tx)

	successCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		externalIdFull := uuid.New()
		externalIdCompressed := uuid.New()
		previousMediaId, err := m.rpcClient.GetUserProfileImageIdRpc(userId)
		if err != nil {
			errCh <- err
			return
		}
		var mediaId int32
		//check if user has a previous profile media
		if previousMediaId > 0 {
			mediaId = previousMediaId
			externalIds, err := txQuries.GetExternalIdsById(timeoutCtx, previousMediaId)
			if err != nil {
				errCh <- err
				return
			}
			msg := rabbitmq_producer.ExternalIdDeletedMsg{
				ExternalId: externalIds.ExternalUuidFull,
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				errCh <- err
				return
			}
			err = m.rabbitmqProducer.Publish("media_events", "media.externalId.deleted", msgBytes)
			if err != nil {
				errCh <- err
				return
			}
			msg2 := rabbitmq_producer.ExternalIdDeletedMsg{
				ExternalId: externalIds.ExternalUuidCompressed,
			}
			msgBytes2, err := json.Marshal(msg2)
			if err != nil {
				errCh <- err
				return
			}
			err = m.rabbitmqProducer.Publish("media_events", "media.externalId.deleted", msgBytes2)
			if err != nil {
				errCh <- err
				return
			}
			err = txQuries.UpdateExternalIdsById(timeoutCtx, media_sql.UpdateExternalIdsByIdParams{
				MediaID:                previousMediaId,
				ExternalUuidFull:       externalIdFull,
				ExternalUuidCompressed: externalIdCompressed,
			})
			if err != nil {
				errCh <- err
				return
			}
		} else {
			newId, err := txQuries.CreateMedia(timeoutCtx, media_sql.CreateMediaParams{
				ExternalUuidFull:       externalIdFull,
				ExternalUuidCompressed: externalIdCompressed,
				UserID:                 userId,
				CompressionStatus:      "started",
				IsActive:               true,
			})
			if err != nil {
				errCh <- err
				return
			}
			err = m.rpcClient.UpdateUserProfileImage(rpc_client.ProfileImageUpdateInput{
				UserId:  userId,
				MediaId: newId,
			})
			if err != nil {
				errCh <- err
				return
			}
			mediaId = newId
		}

		// upload image to minio
		_, err = m.minioClient.PutObject(timeoutCtx, m.config.MinioBucketName, externalIdFull.String(), image, imageHeader.Size, minio.PutObjectOptions{
			ContentType: imageHeader.Header.Get("Content-Type"),
		})
		if err != nil {
			errCh <- err
			return
		}
		msg := rabbitmq_producer.MediaUploadedMsg{
			MediaId:              mediaId,
			ExternalIdFull:       externalIdFull,
			ExternalIdCompressed: externalIdCompressed,
			ContentType:          imageHeader.Header.Get("Content-Type"),
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			errCh <- err
			return
		}
		err = m.rabbitmqProducer.Publish("media_events", "media.uploaded", msgBytes)
		if err != nil {
			errCh <- err
			return
		}

		successCh <- true
	}()

	select {
	case <-successCh:
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		tx.Rollback()
		return err
	case <-timeoutCtx.Done():
		tx.Rollback()
		return timeoutCtx.Err()
	}
}

func (m *MediaService) UpdateCompressionStatus(ctx context.Context, mediaId int32, status string) error {
	err := m.mediaQueries.UpdateCompressionStatus(ctx, media_sql.UpdateCompressionStatusParams{
		MediaID:           mediaId,
		CompressionStatus: status,
	})
	if err != nil {
		return err
	}
	return nil
}
