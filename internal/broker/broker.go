package broker

import (
	"fmt"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/streadway/amqp"
)

var _broker *Broker

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewBroker(conf *config.RabbitMQConfig) (*Broker, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s/", conf.User, conf.Password, conf.Host)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %s", err)
	}

	_broker = &Broker{
		conn: conn,
		ch:   ch,
	}

	return _broker, nil
}

func GetBroker() *Broker {
	return _broker
}

func (b *Broker) Publish(exchange, key, contentType string, body *[]byte) error {
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
		return fmt.Errorf("failed to publish message: %s", err)
	}

	return nil
}

func (b *Broker) Consume(exchange, kind, key string) (<-chan amqp.Delivery, error) {
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
		return nil, fmt.Errorf("failed to declare an exchange: %s", err)
	}

	q, err := b.ch.QueueDeclare("", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %s", err)
	}

	err = b.ch.QueueBind(q.Name, key, exchange, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind a queue: %s", err)
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
		return nil, fmt.Errorf("failed to consume messages: %s", err)
	}

	return messages, nil
}

func (b *Broker) Close() error {
	err := b.ch.Close()
	if err != nil {
		return fmt.Errorf("unable to close channel: %s", err)
	}

	err = b.conn.Close()
	if err != nil {
		return fmt.Errorf("unable to close connection: %s", err)
	}

	return nil
}