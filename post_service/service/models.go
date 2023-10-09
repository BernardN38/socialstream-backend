package service

import (
	"mime/multipart"

	"github.com/go-playground/validator/v10"
)

type CreatePostInput struct {
	UserId    int32  `json:"userId" validate:"required"`
	Username  string `json:"username", validate:"required"`
	Body      string `json:"body" validate:"required"`
	MediaId   int32  `json:"mediaId"`
	Media     multipart.File
	MediaType string
	MediaSize int64
}

func Validate(input interface{}) error {
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		return err
	}
	return nil
}
