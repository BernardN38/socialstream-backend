package application

import (
	"os"
	"testing"
)

// happy path
func TestGetEnvConfig(t *testing.T) {
	port := ":8080"
	postgresDsn := "user:user@localhost"
	os.Setenv("port", port)
	os.Setenv("postgresDsn", postgresDsn)
	config, err := getEnvConfig()
	if err != nil {
		t.Error(err)
	}
	if config.Port != port {
		t.Error("invalid port")
	}
	if config.PostgresDsn != postgresDsn {
		t.Error("invalid postgresDsn")
	}
}

// empty config should return error
func TestGetEnvConfigEmpty(t *testing.T) {
	port := ""
	postgresDsn := ""
	os.Setenv("port", port)
	os.Setenv("postgresDsn", postgresDsn)
	_, err := getEnvConfig()
	if err == nil {
		t.Error("config invalid, expected error: got nil")
	}
}
