package service

import "github.com/go-playground/validator/v10"

type CreateUserInput struct {
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

func Validate(input interface{}) error {
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
