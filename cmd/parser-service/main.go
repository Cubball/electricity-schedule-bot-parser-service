package main

import (
	"electricity-schedule-bot/parser-service/internal/config"
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/publisher"
	"electricity-schedule-bot/parser-service/internal/runner"
	"fmt"
)

func main() {
	config, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %q", err)
		return
	}

	fetcher := fetcher.New(fetcher.FetcherConfig{
		Url: config.WebPageUrl,
	})
	publisher, err := publisher.New(publisher.PublisherConfig{
		RabbitMQUrl:  config.RabbitMqUrl,
		ExchangeName: config.ExchangeName,
		RoutingKey:   config.RoutingKey,
	})
	if err != nil {
		fmt.Printf("failed to init publisher: %q", err)
		return
	}

	runner := runner.New(runner.RunnerConfig{
		FetchInterval:  config.FetchInterval,
		RunImmediately: config.RunImmediately,
		Fetcher:        fetcher,
		Publisher:      publisher,
	})
	runner.Run()
	runner.Wait()
}
