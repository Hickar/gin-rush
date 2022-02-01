package broker

import (
	"fmt"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/streadway/amqp"
)

type Broker interface {
	Publish(string, string, string, *[]byte) error
	Consume(string, string, string) (<-chan amqp.Delivery, error)
	Close() error
}

type RabbitMQBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewBroker(conf *config.RabbitMQConfig) (*RabbitMQBroker, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s/", conf.User, conf.Password, conf.Host)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQBroker{conn: conn, ch: ch}, nil
}

func (b *RabbitMQBroker) Publish(exchange, key, contentType string, body *[]byte) error {
	err := b.ch.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        *body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (b *RabbitMQBroker) Consume(exchange, kind, key string) (<-chan amqp.Delivery, error) {
	err := b.ch.ExchangeDeclare(
		exchange,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	q, err := b.ch.QueueDeclare("", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = b.ch.QueueBind(q.Name, key, exchange, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind a queue: %w", err)
	}

	messages, err := b.ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	return messages, nil
}

func (b *RabbitMQBroker) Close() error {
	if err := b.ch.Close(); err != nil {
		return fmt.Errorf("unable to close channel: %w", err)
	}

	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("unable to close connection: %w", err)
	}

	return nil
}
