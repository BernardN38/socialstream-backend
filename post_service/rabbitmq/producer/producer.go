package rabbitmq_producer

import (
	"github.com/streadway/amqp"
)

type RabbitMQProducerInterface interface {
	Publish(string, []byte) error
	Close() error
}

// RabbitMQProducer represents a RabbitMQ message producer.
type RabbitMQProducer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

func NewRabbitMQProducer(conn *amqp.Connection, exchangeName string) (*RabbitMQProducer, error) {
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = channel.ExchangeDeclare(
		exchangeName,
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

	return &RabbitMQProducer{
		conn:     conn,
		channel:  channel,
		exchange: exchangeName,
	}, nil
}

func (p *RabbitMQProducer) Close() error {
	if p.channel != nil {
		err := p.channel.Close()
		if err != nil {
			return err
		}
	}

	if p.conn != nil {
		return p.conn.Close()
	}

	return nil
}

// Publish sends a message to the RabbitMQ exchange with the given routing key.
func (p *RabbitMQProducer) Publish(exchange string, topic string, message []byte) error {
	err := p.channel.Publish(
		exchange, // exchange: the name of the topic exchange
		topic,    // routing key: the topic or event type, e.g., "create.user" or "update.user"
		false,    // mandatory: false means the message can be silently dropped if no queue is bound to the exchange
		false,    // immediate: false means the message can be queued if it can't be delivered immediately
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)

	return err
}
