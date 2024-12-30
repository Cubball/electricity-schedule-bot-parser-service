package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

func (p *Publisher) Publish(ctx context.Context, obj any) error {
	body, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal an object to json: %w", err)
	}

	traceId, ok := ctx.Value("traceId").(string)
	if !ok {
		traceId = uuid.NewString()
		slog.WarnContext(ctx, "failed to get the `traceId` from the context, generating a new one", "traceId", traceId)
	}

	headers := amqp.Table{
		traceIdHeaderName: traceId,
	}
    slog.DebugContext(ctx, "will publish a message to rmq", "content", body)
	err = p.channel.Publish(p.exchangeName, p.routingKey, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        body,
		Headers:     headers,
	})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

    slog.DebugContext(ctx, "published a message to rmq")
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
    slog.Info("connected to rmq")
	return nil
}

func (p *Publisher) handleReconnect() {
	go func() {
		for {
			err := <-p.closeChannel
			if err == nil {
				break
			}

            slog.Warn("disconnected from rmq")
			for {
                slog.Info("trying to reconnect to rmq")
				err := p.connect()
				if err == nil {
					break
				}

				time.Sleep(time.Second * reconnectWaitInSeconds)
			}
		}
	}()
}
