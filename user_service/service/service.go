package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mime/multipart"
	"time"

	"github.com/BernardN38/flutter-backend/user_service/sql/users"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UserService struct {
	userDb       *sql.DB
	userDbQuries *users.Queries
	minioClient  *minio.Client
	config       *UserServiceConfig
}
type UserServiceConfig struct {
	MinioBucketName string
}

func New(userDb *sql.DB, minioClient *minio.Client, config UserServiceConfig) *UserService {
	userDbQueries := users.New(userDb)
	bucketName := config.MinioBucketName
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Fatalln(err)
	}

	if !exists {
		// Create the bucket
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Bucket '%s' created successfully\n", bucketName)
	} else {
		log.Printf("Bucket '%s' already exists\n", bucketName)
	}
	return &UserService{
		userDb:       userDb,
		userDbQuries: userDbQueries,
		minioClient:  minioClient,
		config:       &config,
	}
}

func (u *UserService) CreateUser(ctx context.Context, createUserInput CreateUserInput) error {
	_, err := u.userDbQuries.CreateUser(ctx, users.CreateUserParams{
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
	// Create a new context with a timeout of 200 ms
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	newImageId := uuid.New()
	tx, err := u.userDb.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQuries := u.userDbQuries.WithTx(tx)
	err = txQuries.UpdateUserProfileImage(ctx, users.UpdateUserProfileImageParams{
		ID: userId,
		ProfileImageID: uuid.NullUUID{
			UUID:  newImageId,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	infoCh := make(chan minio.UploadInfo)
	errCh := make(chan error)

	go func() {
		info, err := u.minioClient.PutObject(ctx, u.config.MinioBucketName, newImageId.String(), image, imageHeader.Size, minio.PutObjectOptions{
			ContentType: imageHeader.Header.Get("Content-Type"),
		})
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
