package rabbitmq_consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/BernardN38/socialstream-backend/authentication_service/service"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	authService *service.AuthSerice
	conn        *amqp.Connection
	channel     *amqp.Channel
	queue       string
}

func NewRabbitMQConsumer(conn *amqp.Connection, queueName string, authService *service.AuthSerice) (*RabbitMQConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		"user_events",
		"topic",
		true,
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

	err = channel.QueueBind(queue.Name, "user.deleted", "user_events", false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConsumer{
		conn:        conn,
		channel:     channel,
		queue:       queue.Name,
		authService: authService,
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
			case "user.deleted":
				var deleteUserReq map[string]int32
				err := json.Unmarshal(msg.Body, &deleteUserReq)
				if err != nil {
					log.Println(err)
					return
				}
				c.authService.DeleteUser(ctx, deleteUserReq["userId"])
				log.Println("auth service deleted user with id: ", deleteUserReq["userId"])
			default:
				log.Println("did not recognize topic:", msg.RoutingKey)
			}
		}
	}()
	<-forever
	return nil
}
