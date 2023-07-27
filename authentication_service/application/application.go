package application

import (
	"database/sql"
	"embed"
	"log"
	"net/http"

	"github.com/BernardN38/flutter-backend/handler"
	"github.com/BernardN38/flutter-backend/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Application struct {
	server *server
}
type server struct {
	router *chi.Mux
	port   string
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New() *Application {
	//get env configuration
	config, err := getEnvConfig()
	if err != nil {
		log.Fatal(err)
	}

	//connect to postgres db
	db, err := sql.Open("postgres", config.PostgresDsn)
	if err != nil {
		log.Fatal(err)
	}

	//run db migrations
	err = RunDatabaseMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	//init service layer
	authService := service.New(db)

	//init handler, inject service
	handler := handler.NewHandler(authService)

	//init server, inject handler & confid
	server := NewServer(handler, config)

	return &Application{
		server: server,
	}
}

func (a *Application) Run() {
	//start server
	log.Printf("listening on port %s", a.server.port)
	log.Fatal(http.ListenAndServe(a.server.port, a.server.router))
}

func NewServer(handler *handler.Handler, config *config) *server {
	r := SetupRouter(handler)
	return &server{
		router: r,
		port:   config.Port,
	}
}

func RunDatabaseMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil
}
