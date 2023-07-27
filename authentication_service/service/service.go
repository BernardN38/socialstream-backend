package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/BernardN38/flutter-backend/sql/users"
	"github.com/lib/pq"
)

type AuthSerice struct {
	authDb       *sql.DB
	authDbQuries *users.Queries
}

func New(authDb *sql.DB) *AuthSerice {
	authDbQueries := users.New(authDb)
	return &AuthSerice{
		authDb:       authDb,
		authDbQuries: authDbQueries,
	}
}

func (a *AuthSerice) CreateUser(ctx context.Context, CreateUserInput CreateUserInput) error {
	user, err := a.authDbQuries.CreateUser(ctx, users.CreateUserParams{
		Username: CreateUserInput.Username,
		Password: CreateUserInput.Password,
		Email:    CreateUserInput.Email,
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
	log.Printf("user created, %+v", user)
	return nil
}
