package rabbitmq_consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	userService *service.UserService
	conn        *amqp.Connection
	channel     *amqp.Channel
	queue       string
}

func NewRabbitMQConsumer(conn *amqp.Connection, queueName string, userService *service.UserService) (*RabbitMQConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		"user_events",
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		conn.Close()
		return nil, err
	}
	queue, err := channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.QueueBind(queue.Name, "", "user_events", false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConsumer{
		conn:        conn,
		channel:     channel,
		queue:       queue.Name,
		userService: userService,
	}, nil
}

func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		err := c.channel.Close()
		if err != nil {
			return err
		}
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *RabbitMQConsumer) Consume() error {
	ctx := context.Background()
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			switch msg.RoutingKey {
			case "user.created":
				var createUserInput service.CreateUserInput
				err := json.Unmarshal(msg.Body, &createUserInput)
				if err != nil {
					log.Println(err)
				}
				err = c.userService.CreateUser(ctx, createUserInput)
				if err != nil {
					log.Println(err)
				}
				log.Println("user created: ", createUserInput)
			default:
				log.Println("did not recognize topic:", msg.RoutingKey)
			}
		}
	}()
	<-forever
	return nil
}
