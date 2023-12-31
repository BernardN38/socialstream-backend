package application

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BernardN38/socialstream-backend/authentication_service/handler"
	rabbitmq_consumer "github.com/BernardN38/socialstream-backend/authentication_service/rabbitmq/consumer"
	rabbitmq_producer "github.com/BernardN38/socialstream-backend/authentication_service/rabbitmq/producer"
	"github.com/BernardN38/socialstream-backend/authentication_service/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/streadway/amqp"
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
	rabbitmqConn, err := ConnectRabbitMQWithRetry(config.RabbitUrl, 5, time.Second*10)
	if err != nil {
		log.Fatal(err)
	}
	rabbitmqProducer, err := rabbitmq_producer.NewRabbitMQProducer(rabbitmqConn, "user_events")
	if err != nil {
		log.Fatal(err)
	}
	//init service layer
	authService := service.New(db, rabbitmqProducer)

	// init rabbitmq Consumer and inject userService to handle messages
	rabbitConsumer, err := rabbitmq_consumer.NewRabbitMQConsumer(rabbitmqConn, "auth-service", authService)
	if err != nil {
		log.Fatal(err)
	}
	//run consumer async
	go func(rabbitConsumer *rabbitmq_consumer.RabbitMQConsumer) {
		err := rabbitConsumer.Consume()
		if err != nil {
			log.Fatal(err)
		}
	}(rabbitConsumer)

	tokenManager := jwtauth.New("HS256", []byte(config.JwtSecret), nil)

	//init handler, inject service
	handler := handler.NewHandler(authService, tokenManager)

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

// ConnectWithRetry establishes a connection to RabbitMQ with wait and retry logic.
func ConnectRabbitMQWithRetry(amqpURL string, maxRetries int, retryInterval time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for retries := 0; retries <= maxRetries; retries++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			log.Println("Connected to RabbitMQ successfully.")
			return conn, nil
		}

		log.Printf("Failed to connect to RabbitMQ: %v", err)
		if retries < maxRetries {
			log.Printf("Retrying connection in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d retries", maxRetries)
}
