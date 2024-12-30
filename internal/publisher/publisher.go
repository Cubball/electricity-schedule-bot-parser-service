package publisher

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	traceIdHeaderName      = "X-Trace-Id"
	reconnectWaitInSeconds = 10
	contentType            = "application/json"
)

type PublisherConfig struct {
	RabbitMQUrl  string
	ExchangeName string
	RoutingKey   string
}

type Publisher struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	rabbitMQUrl  string
	exchangeName string
	routingKey   string
	closeChannel chan *amqp.Error
}

func New(config PublisherConfig) (*Publisher, error) {
	publisher := Publisher{
		rabbitMQUrl:  config.RabbitMQUrl,
		exchangeName: config.ExchangeName,
		routingKey:   config.RoutingKey,
	}
	err := publisher.connect()
	if err != nil {
		return nil, err
	}

	publisher.handleReconnect()
	return &publisher, nil
}

func (p *Publisher) Publish(obj any) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal an object to json: %w", err)
	}

	headers := amqp.Table{
		traceIdHeaderName: uuid.New(),
	}
	err = p.channel.Publish(p.exchangeName, p.routingKey, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.connection.Close()
}

func (p *Publisher) connect() error {
	connection, err := amqp.Dial(p.rabbitMQUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to rmq: %w", err)
	}

	p.connection = connection
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	p.channel = channel
	err = channel.ExchangeDeclare(p.exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %w", err)
	}

	p.closeChannel = connection.NotifyClose(make(chan *amqp.Error))
	return nil
}

func (p *Publisher) handleReconnect() {
	go func() {
		for {
			err := <-p.closeChannel
			if err == nil {
				break
			}

			for {
				err := p.connect()
				if err == nil {
					break
				}

				time.Sleep(time.Second * reconnectWaitInSeconds)
			}
		}
	}()
}
