package service

import "github.com/go-playground/validator/v10"

type CreateUserInput struct {
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type LoginUserInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserCreatedMessage struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func Validate(input *interface{}) error {
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
