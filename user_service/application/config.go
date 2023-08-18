package application

import (
	"os"

	"github.com/go-playground/validator/v10"
)

type config struct {
	Port                 string `validate:"required"`
	PostgresDsn          string `validate:"required"`
	JwtSecret            string `validate:"required"`
	RabbitUrl            string `validate:"required"`
	MinioEndpoint        string `validate:"required"`
	MinioAccessKeyID     string `validate:"required"`
	MinioSecretAccessKey string `validate:"required"`
}

func (c *config) Validate() error {
	validator := validator.New()
	err := validator.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

func getEnvConfig() (*config, error) {
	port := os.Getenv("port")
	postgresDsn := os.Getenv("postgresDsn")
	jwtSecret := os.Getenv("jwtSecret")
	rabbitUrl := os.Getenv("rabbitmqUrl")
	minioAccessKeyID := os.Getenv("minioAccessKeyID")
	minioSecretAccessKey := os.Getenv("minioSecretAccessKey")
	minioEndpoint := os.Getenv("minioEndpoint")
	config := config{
		Port:                 port,
		PostgresDsn:          postgresDsn,
		JwtSecret:            jwtSecret,
		RabbitUrl:            rabbitUrl,
		MinioAccessKeyID:     minioAccessKeyID,
		MinioSecretAccessKey: minioSecretAccessKey,
		MinioEndpoint:        minioEndpoint,
	}
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	return &config, config.Validate()
}
