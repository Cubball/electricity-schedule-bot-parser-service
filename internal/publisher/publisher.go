package publisher

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	RabbitMQUrl  string
	ExchangeName string
	RoutingKey   string
}

type Publisher struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	routingKey   string
}

// TODO: retry logic
func New(config *PublisherConfig) (*Publisher, error) {
	connection, err := amqp.Dial(config.RabbitMQUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = channel.ExchangeDeclare(config.ExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	return &Publisher{
		connection:   connection,
		channel:      channel,
		exchangeName: config.ExchangeName,
		routingKey:   config.RoutingKey,
	}, nil
}

func (p *Publisher) Publish(obj any) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal an object to json: %w", err)
	}

	err = p.channel.Publish(p.exchangeName, p.routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
    // TODO: remove me
    fmt.Printf("published: %q", string(body))
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (p *Publisher) Close() {
    p.channel.Close()
    p.connection.Close()
}
