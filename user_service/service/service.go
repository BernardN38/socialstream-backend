package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"time"

	rpc_client "github.com/BernardN38/socialstream-backend/user_service/rpc/client"
	"github.com/BernardN38/socialstream-backend/user_service/sql/users"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UserService struct {
	userDb       *sql.DB
	userDbQuries *users.Queries
	minioClient  *minio.Client
	rpcClient    *rpc_client.RpcClient
	config       *UserServiceConfig
}
type UserServiceConfig struct {
	MinioBucketName string
}

func New(userDb *sql.DB, minioClient *minio.Client, rpcClient *rpc_client.RpcClient, config UserServiceConfig) (*UserService, error) {
	userDbQueries := users.New(userDb)
	err := setup(*minioClient, "user-service")
	if err != nil {
		return nil, err
	}
	return &UserService{
		userDb:       userDb,
		userDbQuries: userDbQueries,
		minioClient:  minioClient,
		rpcClient:    rpcClient,
		config:       &config,
	}, nil
}

func (u *UserService) CreateUser(ctx context.Context, createUserInput CreateUserInput) error {
	_, err := u.userDbQuries.CreateUser(ctx, users.CreateUserParams{
		UserID:    createUserInput.UserId,
		Username:  createUserInput.Username,
		Email:     createUserInput.Email,
		Firstname: createUserInput.FirstName,
		Lastname:  createUserInput.LastName,
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *UserService) GetAllUsers(ctx context.Context) ([]users.User, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel() // Make sure to call cancel to release resources when done
	userChan := make(chan []users.User)
	errChan := make(chan error)

	go func() {
		users, err := u.userDbQuries.ListUsers(ctx)
		if err != nil {
			errChan <- err
		}
		userChan <- users
	}()
	select {
	case <-timeoutCtx.Done():
		return nil, errors.New("get all users timedout")
	case users := <-userChan:
		return users, nil
	case err := <-errChan:
		return nil, err

	}
}

func (u *UserService) GetUser(ctx context.Context, userId int32) (users.User, error) {
	// Create a new context with a timeout of 200 milliseconds
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel() // Make sure to call cancel to release resources when done

	userChan := make(chan users.User)
	errChan := make(chan error)

	go func() {
		user, err := u.userDbQuries.GetUserById(timeoutCtx, userId)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	select {
	case <-timeoutCtx.Done():
		return users.User{}, timeoutCtx.Err()
	case user := <-userChan:
		return user, nil
	case err := <-errChan:
		return users.User{}, err
	}
}
func (u *UserService) UpdateUserProfileImage(ctx context.Context, userId int32, image multipart.File, imageHeader *multipart.FileHeader) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	tx, err := u.userDb.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQuries := u.userDbQuries.WithTx(tx)
	profileImageId, err := txQuries.GetUserProfileImageByUserId(ctx, userId)
	if err != nil {
		return err
	}

	if profileImageId.Valid {
		err := u.rpcClient.DeleteMedia(profileImageId.UUID)
		if err != nil {
			return err
		}
	}
	newImageId := uuid.New()
	err = txQuries.UpdateUserProfileImage(timeoutCtx, users.UpdateUserProfileImageParams{
		UserID: userId,
		ProfileImageID: uuid.NullUUID{
			UUID:  newImageId,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	successCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		imageBytes, err := io.ReadAll(image)
		if err != nil {
			errCh <- err
			return
		}
		err = u.rpcClient.UploadMedia(&rpc_client.ImageUpload{
			ImageData:   imageBytes,
			MediaId:     newImageId,
			ContentType: imageHeader.Header.Get("Content-Type"),
		})
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

func (u *UserService) GetUserProfileImage(ctx context.Context, userId int32) (uuid.UUID, error) {
	mediaId, err := u.userDbQuries.GetUserProfileImageByUserId(ctx, userId)
	if err != nil {
		return uuid.UUID{}, err
	}
	return mediaId.UUID, nil
}
