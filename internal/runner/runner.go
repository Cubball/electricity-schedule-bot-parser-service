package runner

import (
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/parser"
	"electricity-schedule-bot/parser-service/internal/publisher"
	"fmt"
	"time"
)

type Runner struct {
	interval    time.Duration
	fetcher     *fetcher.Fetcher
	publisher   *publisher.Publisher
	stopChannel chan bool
}

func New(interval time.Duration, fetcher *fetcher.Fetcher, publisher *publisher.Publisher) *Runner {
	return &Runner{
		interval:    interval,
		fetcher:     fetcher,
		publisher:   publisher,
		stopChannel: make(chan bool),
	}
}

func (r *Runner) Run() {
	ticker := time.NewTicker(r.interval)
	go func() {
        runImmediately := true // TODO: put in config
        if runImmediately {
            r.run()
        }

		for {
			select {
			case <-r.stopChannel:
				ticker.Stop()
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

// TODO: logging
func (r *Runner) run() {
	webPage, err := r.fetcher.Fetch()
	if err != nil {
		return
	}

	schedule, err := parser.Parse(webPage)
	if err != nil {
        fmt.Printf("error parsing: %q", err)
		return
	}

    fmt.Println("publishing")
	err = r.publisher.Publish(schedule)
	if err != nil {
	}
}
