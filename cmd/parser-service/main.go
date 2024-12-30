package main

import (
	"electricity-schedule-bot/parser-service/internal/config"
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/logger"
	"electricity-schedule-bot/parser-service/internal/publisher"
	"electricity-schedule-bot/parser-service/internal/runner"
	"log/slog"
	"os"
)

func main() {
	config, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		return
	}

	var handler slog.Handler
	if config.IsProduction {
		handler = logger.NewTraceIdHandler(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		handler = logger.NewTraceIdHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	logger := slog.New(handler)
	logger = logger.With("service", "parser-service")
	slog.SetDefault(logger)

	fetcher := fetcher.New(fetcher.FetcherConfig{
		Url: config.WebPageUrl,
	})
	publisher, err := publisher.New(publisher.PublisherConfig{
		RabbitMQUrl:  config.RabbitMqUrl,
		ExchangeName: config.ExchangeName,
		RoutingKey:   config.RoutingKey,
	})
	if err != nil {
		slog.Error("failed to init publisher", "err", err)
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
