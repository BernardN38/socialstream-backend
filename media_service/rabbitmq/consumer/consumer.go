package rabbitmq_comsumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/BernardN38/socialstream-backend/media_service/service"
	"github.com/streadway/amqp"
)

type RabbitMQConsumer struct {
	mediaService *service.MediaService
	conn         *amqp.Connection
	channel      *amqp.Channel
	queue        string
}

func NewRabbitMQConsumer(conn *amqp.Connection, queueName string, mediaService *service.MediaService) (*RabbitMQConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		"media_events",
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

	err = channel.QueueBind(queue.Name, "media.compressed", "media_events", false, nil)
	if err != nil {
		return nil, err
	}
	err = channel.QueueBind(queue.Name, "media.deleted", "media_events", false, nil)
	if err != nil {
		return nil, err
	}
	err = channel.QueueBind(queue.Name, "media.externalId.deleted", "media_events", false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQConsumer{
		conn:         conn,
		channel:      channel,
		queue:        queue.Name,
		mediaService: mediaService,
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
		false,
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
				log.Println("user created: ")
			case "media.externalId.deleted":
				mediaDeletedMsg := ExternalIdDeletedMsg{}
				err := json.Unmarshal(msg.Body, &mediaDeletedMsg)
				if err != nil {
					log.Println(err)
					msg.Nack(false, true)
					continue
				}
				err = c.mediaService.DeleteExternalId(ctx, mediaDeletedMsg.ExternalId)
				if err != nil {
					log.Println(err)
					if time.Since(msg.Timestamp) > 10*time.Second {
						msg.Nack(false, false)
						continue
					}
					msg.Nack(false, true)
					continue
				}
				log.Println("externalId deleted: ", mediaDeletedMsg.ExternalId)
				msg.Ack(true)
			case "media.compressed":
				mediaCompressedMsg := MediaCompressedMsg{}
				err := json.Unmarshal(msg.Body, &mediaCompressedMsg)
				if err != nil {
					log.Println(err)
					continue
				}
				err = c.mediaService.UpdateCompressionStatus(ctx, mediaCompressedMsg.MediaId, "complete")
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("media id compression complete:", mediaCompressedMsg.MediaId)
				msg.Ack(true)
			case "media.deleted":
				mediaDeletedMsg := MediaDeletedMsg{}
				err := json.Unmarshal(msg.Body, &mediaDeletedMsg)
				if err != nil {
					log.Println(err)
					continue
				}
				err = c.mediaService.DeleteMedia(ctx, mediaDeletedMsg.MediaId)
				if err != nil {
					log.Println(err)
					msg.Nack(false, true)
					continue
				}
				msg.Ack(true)
			default:
				log.Println("did not recognize topic:", msg.RoutingKey)
			}
		}
	}()
	<-forever
	return nil
}
