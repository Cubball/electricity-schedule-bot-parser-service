package main

import (
	"electricity-schedule-bot/parser-service/internal/fetcher"
	"electricity-schedule-bot/parser-service/internal/parser"
	"fmt"
)

const webPageUrl string = "https://www.roe.vsei.ua/disconnections"

// const webPageUrl string = "https://google.com"

// TODO: config
func main() {
    fetcher := fetcher.New(webPageUrl)
    webPage, err := fetcher.Fetch()
    if err != nil {
        fmt.Printf("err: %q", err)
        return;
    }

    schedule, err := parser.Parse(webPage)
    if err != nil {
        fmt.Printf("err: %q", err)
        return;
    }

    fmt.Printf("date: %q, len: %d", schedule.FetchTime, len(schedule.Entries))
}
