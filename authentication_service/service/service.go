package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/BernardN38/flutter-backend/authentication_service/rabbitmq"
	"github.com/BernardN38/flutter-backend/authentication_service/sql/users"
	"github.com/lib/pq"
)

type AuthSerice struct {
	authDb           *sql.DB
	authDbQuries     *users.Queries
	rabbitmqProducer rabbitmq.RabbitMQProducerInterface
}

func New(authDb *sql.DB, rabbitmqProducer rabbitmq.RabbitMQProducerInterface) *AuthSerice {
	authDbQueries := users.New(authDb)
	return &AuthSerice{
		authDb:           authDb,
		authDbQuries:     authDbQueries,
		rabbitmqProducer: rabbitmqProducer,
	}
}
func (a *AuthSerice) GetAllUsers(ctx context.Context) ([]users.GetAllUsersRow, error) {
	users, err := a.authDbQuries.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (a *AuthSerice) CreateUser(ctx context.Context, createUserInput CreateUserInput, role string) error {
	if role == "" {
		role = "user"
	}
	user, err := a.authDbQuries.CreateUser(ctx, users.CreateUserParams{
		Username: createUserInput.Username,
		Password: createUserInput.Password,
		Email:    createUserInput.Email,
		Role:     role,
	})
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			switch e.Code {
			case "23505":
				log.Println(err)
				return errors.New("username or email already taken")
			default:
				log.Println(err)
				return errors.New("unkown error databse error")
			}
		default:
			log.Println(err)
			return errors.New("database error")
		}
	}
	message, err := json.Marshal(UserCreatedMessage{Username: user.Username, UserId: user.ID, Email: user.Email, FirstName: createUserInput.FirstName, LastName: createUserInput.LastName})
	if err != nil {
		return err
	}
	a.rabbitmqProducer.Publish("user.created", message)
	log.Printf("user created, %+v", user)
	return nil
}

func (a *AuthSerice) LoginUser(ctx context.Context, loginUserInput LoginUserInput) (users.User, error) {
	user, err := a.authDbQuries.GetUserByUsername(ctx, loginUserInput.Username)
	if err != nil {
		return users.User{}, err
	}
	if user.Password != loginUserInput.Password {
		return users.User{}, errors.New("unathorized")
	}
	user.Password = ""
	return user, nil
}
