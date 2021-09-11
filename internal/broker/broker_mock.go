package broker

import "github.com/streadway/amqp"

type brokerMock struct {
	broker
}

func NewBrokerMock() (Broker, error) {
	_broker = &brokerMock{}

	return _broker, nil
}

func (b *brokerMock) Publish(exchange, key, contentType string, body *[]byte) error {
	return nil
}

func (b *brokerMock) Consume(exchange, kind, key string) (<-chan amqp.Delivery, error) {
	return nil, nil
}

func (b *brokerMock) Close() error {
	return nil
}