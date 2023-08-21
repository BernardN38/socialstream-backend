package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	rabbitmq_producer "github.com/BernardN38/flutter-backend/media_service/rabbitmq/producer"
	rpc_client "github.com/BernardN38/flutter-backend/media_service/rpc/client"
	media_sql "github.com/BernardN38/flutter-backend/media_service/sql/media"
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
type RpcImageUpload struct {
	MediaData   []byte
	MediaId     uuid.UUID
	ContentType string
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

func (m *MediaService) GetImage(ctx context.Context, imageId uuid.UUID) (*minio.Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	respCh := make(chan *minio.Object)
	errCh := make(chan error)
	go func() {
		object, err := m.minioClient.GetObject(ctx, "media-service", imageId.String(), minio.GetObjectOptions{})
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
func (m *MediaService) UploadMedia(ctx context.Context, payload RpcImageUpload) error {
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	tx, err := m.mediaDb.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQuries := m.mediaQueries.WithTx(tx)
	_, err = txQuries.CreateMedia(ctx, media_sql.CreateMediaParams{
		MediaID: payload.MediaId,
		OwnerID: 1,
	})
	if err != nil {
		return err
	}

	infoCh := make(chan minio.UploadInfo)
	errCh := make(chan error)

	go func() {
		info, err := m.minioClient.PutObject(ctx, m.config.MinioBucketName, payload.MediaId.String(), bytes.NewReader(payload.MediaData), int64(len(payload.MediaData)), minio.PutObjectOptions{
			ContentType: payload.ContentType,
		})
		if err != nil {
			errCh <- err
			return
		}
		msg, err := json.Marshal(struct {
			MediaId     string `json:"mediaId"`
			ContentType string `json:"contentType"`
		}{
			MediaId:     payload.MediaId.String(),
			ContentType: payload.ContentType,
		})
		if err != nil {
			errCh <- errors.New("error publishing message")
			return
		}
		m.rabbitmqProducer.Publish("media.uploaded", msg)
		infoCh <- info
	}()

	select {
	case info := <-infoCh:
		err = tx.Commit()
		if err != nil {
			return err
		}
		log.Println(info)
		return nil
	case err := <-errCh:
		tx.Rollback()
		return err
	case <-ctx.Done():
		tx.Rollback()
		return ctx.Err()
	}
}

func (m *MediaService) DeleteMedia(ctx context.Context, mediaId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	tx, err := m.mediaDb.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQuries := m.mediaQueries.WithTx(tx)
	err = txQuries.DeleteMedia(ctx, mediaId)
	if err != nil {
		return err
	}

	successCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		err := m.minioClient.RemoveObject(ctx, m.config.MinioBucketName, mediaId.String(), minio.RemoveObjectOptions{})
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
		log.Println(err)
		return err
	case <-ctx.Done():
		tx.Rollback()
		return ctx.Err()
	}
}

func (m *MediaService) GetUserProfileImage(reqCtx context.Context, userId int32) (*minio.Object, error) {
	ctx, cancel := context.WithTimeout(reqCtx, 500*time.Millisecond)
	defer cancel()
	respCh := make(chan *minio.Object)
	errCh := make(chan error)
	mediaId, err := m.rpcClient.GetUserProfileImageIdRpc(userId)
	if err != nil {
		cancel()
		return nil, err
	}
	go func(context.Context) {
		object, err := m.minioClient.GetObject(reqCtx, "media-service", mediaId.String(), minio.GetObjectOptions{})
		if err != nil {
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
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// testing compression
// func compressAndLogSizes(inputData []byte, quality int) (*bytes.Buffer, error) {
// 	startTime := time.Now().UnixMilli()
// 	// Create an image.Image from the byte slice
// 	img, _, err := image.Decode(bytes.NewReader(inputData))
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Compress the image
// 	var compressedBufferBefore bytes.Buffer
// 	err = jpeg.Encode(&compressedBufferBefore, img, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the size before compression
// 	sizeBefore := len(compressedBufferBefore.Bytes())

// 	// Set the compression options
// 	var opt jpeg.Options
// 	opt.Quality = quality

// 	// Compress the image with specified quality
// 	var compressedBufferAfter bytes.Buffer
// 	err = jpeg.Encode(&compressedBufferAfter, img, &opt)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the size after compression
// 	sizeAfter := len(compressedBufferAfter.Bytes())
// 	endTime := time.Now().UnixMilli()
// 	fmt.Println("runtime: ", endTime-startTime, "size before: ", sizeBefore, "size after: ", sizeAfter)
// 	return &compressedBufferAfter, nil
// }
