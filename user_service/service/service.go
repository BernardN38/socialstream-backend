package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	rabbitmq_producer "github.com/BernardN38/socialstream-backend/user_service/rabbitmq/producer"
	rpc_client "github.com/BernardN38/socialstream-backend/user_service/rpc/client"
	"github.com/BernardN38/socialstream-backend/user_service/sql/users"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	userDb           *sql.DB
	userDbQuries     *users.Queries
	redisClient      *redis.Client
	minioClient      *minio.Client
	rpcClient        *rpc_client.RpcClient
	rabbitmqPorducer *rabbitmq_producer.RabbitMQProducer
	config           *UserServiceConfig
}
type UserServiceConfig struct {
	MinioBucketName string
}

func New(userDb *sql.DB, redisClient *redis.Client, minioClient *minio.Client, rpcClient *rpc_client.RpcClient, rabbitmqProducer *rabbitmq_producer.RabbitMQProducer, config UserServiceConfig) (*UserService, error) {
	userDbQueries := users.New(userDb)
	err := setup(*minioClient, "user-service")
	if err != nil {
		return nil, err
	}
	return &UserService{
		userDb:           userDb,
		userDbQuries:     userDbQueries,
		redisClient:      redisClient,
		minioClient:      minioClient,
		rpcClient:        rpcClient,
		rabbitmqPorducer: rabbitmqProducer,
		config:           &config,
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

func (u *UserService) UpdateUser(ctx context.Context, userId int32, updateUserInput UpdateUserInput) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	errCh := make(chan error)
	successCh := make(chan interface{})
	go func() {
		err := u.userDbQuries.UpdateUser(timeoutCtx, users.UpdateUserParams{
			UserID:  userId,
			Column2: updateUserInput.Username,
			Column3: updateUserInput.FirstName,
			Column4: updateUserInput.LastName,
		})
		if err != nil {
			errCh <- err
			return
		}
		successCh <- struct{}{}
	}()
	select {
	case err := <-errCh:
		return err
	case <-successCh:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

func (u *UserService) DeleteUser(ctx context.Context, userId int32) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	tx, err := u.userDb.BeginTx(timeoutCtx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQuries := u.userDbQuries.WithTx(tx)
	errCh := make(chan error)
	successCh := make(chan interface{})
	go func() {
		err := txQuries.DeleteUser(timeoutCtx, userId)
		if err != nil {
			errCh <- err
			return
		}
		msgBytes, err := json.Marshal(rabbitmq_producer.UserDeletedMsg{
			UserId: userId,
		})
		if err != nil {
			errCh <- err
			return
		}
		u.rabbitmqPorducer.Publish("user.deleted", msgBytes)
		successCh <- struct{}{}
	}()
	select {
	case err := <-errCh:
		tx.Rollback()
		return err
	case <-successCh:
		tx.Commit()
		return nil
	case <-timeoutCtx.Done():
		tx.Rollback()
		return timeoutCtx.Err()
	}
}

func (u *UserService) GetUserProfileImage(ctx context.Context, userId int32) (int32, error) {
	mediaId, err := u.userDbQuries.GetUserProfileImageByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}
	return mediaId.Int32, nil
}

func (u *UserService) UpdateUserProfileImageId(ctx context.Context, userId int32, mediaId int32) error {
	err := u.userDbQuries.UpdateUserProfileImage(ctx, users.UpdateUserProfileImageParams{
		UserID: userId,
		ProfileImageID: sql.NullInt32{
			Int32: mediaId,
			Valid: mediaId > 0,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
