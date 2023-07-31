package handler

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BernardN38/flutter-backend/authentication_service/service"
	"github.com/go-chi/jwtauth/v5"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	instance testcontainers.Container
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewTestDatabase(t *testing.T) *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	defer cancel()
	req := testcontainers.ContainerRequest{
		Name:         "postgrestestcontainer",
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForLog("LOG:  database system is ready to accept connections"),
		// Privileged: true,
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{

		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	require.NoError(t, err)
	return &TestDatabase{
		instance: postgres,
	}
}

func (db *TestDatabase) Port(t *testing.T) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := db.instance.MappedPort(ctx, "5432")
	require.NoError(t, err)
	return p.Int()
}

func (db *TestDatabase) ConnectionString(t *testing.T) string {
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/postgres?sslmode=disable", db.Port(t))
}

func (db *TestDatabase) Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	require.NoError(t, db.instance.Terminate(ctx))
}

func Setup(t *testing.T) (*httptest.Server, service.AuthSerice, *sql.DB) {
	db := SetupDatabase(t)
	authService := service.New(db, nil)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewHandler(authService, tokenAuth)
	server := httptest.NewServer(http.HandlerFunc(h.LoginUser))
	return server, *authService, db
}
func TearDown(t *testing.T, db *sql.DB) {
	// time.Sleep(time.Second * 10)
	if err := goose.Reset(db, "migrations"); err != nil {
		t.Error(err)
	}
}

func SetupDatabase(t *testing.T) *sql.DB {
	testDatabase := NewTestDatabase(t)
	db, err := sql.Open("postgres", testDatabase.ConnectionString(t))
	if err != nil {
		t.Error(err)
	}
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		t.Error(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		t.Error(err)
	}
	if err := db.Ping(); err != nil {
		t.Error(err)
	}
	return db
}
func TestLoginUser(t *testing.T) {
	ctx := context.Background()
	server, authService, db := Setup(t)
	//set up user in auth database
	username := "testuser"
	password := "testpassword"
	email := "testEmail@test.com"
	err := authService.CreateUser(ctx, service.CreateUserInput{
		Username: username,
		Password: password,
		Email:    email,
	})
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	client := &http.Client{}
	body := []byte(`{"username": "testuser", "password": "testpassword"}`)
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Error("bad response status code, expected 200 got:", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	}
	var bodyJson map[string]int32
	if err := json.Unmarshal(bodyBytes, &bodyJson); err != nil {
		t.Error(err)
	}
	if bodyJson["userId"] != 1 {
		t.Error(`invalid json response, expected {"userId":1}, got : `, string(bodyBytes))
	}

	//chcek response sets cookie
	jwt := resp.Header.Get("Set-Cookie")
	if len(jwt) < 1 {
		t.Error("empty set cookie response")
	}

	//check jwt is set in set cookie header
	jwtIncluded := strings.Contains(jwt, "jwt=")
	if !jwtIncluded {
		t.Error("jwt not set in set cookie header")
	}
	TearDown(t, db)
}
