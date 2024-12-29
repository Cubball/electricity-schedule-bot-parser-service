package main

import (
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/publisher"
	"electricity-schedule-bot/parser-service/internal/runner"
	"fmt"
	"time"
)

const webPageUrl string = "https://www.roe.vsei.ua/disconnections"
const rabbitMqUrl string = "amqp://guest:guest@localhost:5672"

// const webPageUrl string = "https://google.com"
func main() {
	fetcher := fetcher.New(webPageUrl)
	publisher, err := publisher.New(&publisher.PublisherConfig{
		RabbitMQUrl:  rabbitMqUrl,
		ExchangeName: "schedule.topic",
		RoutingKey:   "schedule.parsed",
	})
    if err != nil {
        fmt.Printf("failed to init publisher: %q", err)
        return
    }

    runner := runner.New(time.Hour, fetcher, publisher)
    runner.Run()
    runner.Wait()
}
