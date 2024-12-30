package runner

import (
	"context"
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/logger"
	"electricity-schedule-bot/parser-service/internal/parser"
	"electricity-schedule-bot/parser-service/internal/publisher"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Runner struct {
	fetchInterval  time.Duration
	fetcher        *fetcher.Fetcher
	publisher      *publisher.Publisher
	stopChannel    chan bool
	runImmediately bool
}

type RunnerConfig struct {
	FetchInterval  time.Duration
	Fetcher        *fetcher.Fetcher
	Publisher      *publisher.Publisher
	RunImmediately bool
}

func New(config RunnerConfig) *Runner {
	return &Runner{
		fetchInterval:  config.FetchInterval,
		fetcher:        config.Fetcher,
		publisher:      config.Publisher,
		runImmediately: config.RunImmediately,
		stopChannel:    make(chan bool),
	}
}

func (r *Runner) Run() {
	slog.Info("the runner has started")
	ticker := time.NewTicker(r.fetchInterval)
	go func() {
		if r.runImmediately {
			slog.Info("executing the first run immediately")
			r.run()
		}

		for {
			select {
			case <-r.stopChannel:
				ticker.Stop()
				slog.Info("the runner has stopped")
				break
			case <-ticker.C:
				r.run()
			}
		}
	}()
}

func (r *Runner) Wait() {
	<-r.stopChannel
}

func (r *Runner) Stop() {
	go func() {
		r.stopChannel <- true
		close(r.stopChannel)
		r.publisher.Close()
	}()
}

func (r *Runner) run() {
	ctx := context.WithValue(context.Background(), logger.TraceIdContextKey, uuid.NewString())
	slog.InfoContext(ctx, "executing the run")
	webPage, err := r.fetcher.Fetch(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error while fetching", "err", err)
		return
	}

	schedule, err := parser.Parse(ctx, webPage)
	if err != nil {
		slog.ErrorContext(ctx, "error while parsing", "err", err)
		return
	}

	err = r.publisher.Publish(ctx, schedule)
	if err != nil {
		slog.ErrorContext(ctx, "error while publishing", "err", err)
	}

	slog.InfoContext(ctx, "the run has finished")
}
