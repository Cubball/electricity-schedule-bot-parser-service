package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultWebPageUrl             = "https://www.roe.vsei.ua/disconnections"
	defaultRabbitMqUrl            = "amqp://guest:guest@localhost:5672"
	defaultFetchIntervalInMinutes = "60"
	defaultExchangeName           = "schedule.topic"
	defaultRoutingKey             = "schedule.parsed"
	defaultRunImmediately         = "true"
)

type Config struct {
	WebPageUrl     string
	RabbitMqUrl    string
	RoutingKey     string
	ExchangeName   string
	RunImmediately bool
	FetchInterval  time.Duration
}

func Load() (*Config, error) {
	fetchIntervalStr := getEnvOrDefault("FETCH_INTERVAL_IN_MINUTES", defaultFetchIntervalInMinutes)
	fetchIntervalInMinutes, err := strconv.Atoi(fetchIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `FETCH_INTERVAL_IN_MINUTES`: %w", err)
	}

    if fetchIntervalInMinutes <= 0 {
        return nil, fmt.Errorf("`FETCH_INTERVAL_IN_MINUTES` should be greater than 0, got %d", fetchIntervalInMinutes)
    }

	runImmediatelyStr := getEnvOrDefault("RUN_IMMEDIATELY", defaultRunImmediately)
	runImmediately, err := strconv.ParseBool(runImmediatelyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `RUN_IMMEDIATELY`: %w", err)
	}

	return &Config{
		WebPageUrl:     getEnvOrDefault("WEB_PAGE_URL", defaultWebPageUrl),
		RabbitMqUrl:    getEnvOrDefault("RABBITMQ_URL", defaultRabbitMqUrl),
		RoutingKey:     getEnvOrDefault("ROUTING_KEY", defaultRoutingKey),
		ExchangeName:   getEnvOrDefault("EXCHANGE_NAME", defaultExchangeName),
		RunImmediately: runImmediately,
		FetchInterval:  time.Minute * time.Duration(fetchIntervalInMinutes),
	}, nil
}

func getEnvOrDefault(envVarName, defaultValue string) string {
	fromEnv, ok := os.LookupEnv(envVarName)
	if ok {
		return fromEnv
	}

	return defaultValue
}
