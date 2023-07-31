package service

import (
	"database/sql"

	"github.com/BernardN38/flutter-backend/user_service/sql/users"
)

type UserSerice struct {
	userDb       *sql.DB
	userDbQuries *users.Queries
}

func New(userDb *sql.DB) *UserSerice {
	userDbQueries := users.New(userDb)
	return &UserSerice{
		userDb:       userDb,
		userDbQuries: userDbQueries,
	}
}
