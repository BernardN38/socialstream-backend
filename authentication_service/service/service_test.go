package service

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
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
func TearDown(t *testing.T, db *sql.DB) {
	// time.Sleep(time.Second * 10)
	if err := goose.Reset(db, "migrations"); err != nil {
		t.Error(err)
	}
}

// test happy path
func TestCreateuser(t *testing.T) {
	ctx := context.Background()

	db := SetupDatabase(t)
	authService := New(db)

	user1 := CreateUserInput{
		Username: "testUsername",
		Email:    "testEmail",
		Password: "testPassword",
	}
	err := authService.CreateUser(ctx, user1)
	if err != nil {
		t.Error(err)
	}
	TearDown(t, db)
}

// test duplicate user
func TestDuplicateUser(t *testing.T) {
	ctx := context.Background()

	db := SetupDatabase(t)
	authService := New(db)

	user1 := CreateUserInput{
		Username: "testUsername",
		Email:    "testEmail@gmail.com",
		Password: "testPassword",
	}
	err := authService.CreateUser(ctx, user1)
	if err != nil {
		t.Error(err)
	}
	err = authService.CreateUser(ctx, user1)
	if err == nil {
		t.Error("duplicate username or email allowed without error")
	}
	if err.Error() != "username or email already taken" {
		t.Error("incorrect error message when duplicate user creation attempted")
	}
	TearDown(t, db)
}

// test invalid input to create user
func TestInvalidCreatUserInput(t *testing.T) {
	testCases := []CreateUserInput{
		// empty username
		{
			Username: "",
			Email:    "testEmail2@test.com",
			Password: "testPassword2",
		},
		// empty email
		{
			Username: "testUsername2",
			Email:    "",
			Password: "testPassword2",
		},
		// empty password
		{
			Username: "testUsername2",
			Email:    "testEmail2@test.com",
			Password: "",
		},
	}
	ctx := context.Background()
	db := SetupDatabase(t)
	authService := New(db)

	for _, v := range testCases {
		err := authService.CreateUser(ctx, v)
		if err == nil {
			t.Error(err)
		}
	}

	TearDown(t, db)
}
