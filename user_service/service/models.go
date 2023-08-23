package service

import "github.com/go-playground/validator/v10"

type CreateUserInput struct {
	UserId    int32  `json:"userId" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}
type UpdateUserInput struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func Validate(input interface{}) error {
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
