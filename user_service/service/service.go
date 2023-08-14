package service

import (
	"context"
	"database/sql"

	"github.com/BernardN38/flutter-backend/user_service/sql/users"
)

type UserService struct {
	userDb       *sql.DB
	userDbQuries *users.Queries
}

func New(userDb *sql.DB) *UserService {
	userDbQueries := users.New(userDb)
	return &UserService{
		userDb:       userDb,
		userDbQuries: userDbQueries,
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
	users, err := u.userDbQuries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserService) GetUser(ctx context.Context, userId int32) (users.User, error) {
	user, err := u.userDbQuries.GetUserById(ctx, userId)
	if err != nil {
		return users.User{}, err
	}
	return user, nil
}
