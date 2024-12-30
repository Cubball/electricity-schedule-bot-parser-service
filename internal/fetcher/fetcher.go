package fetcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	timeoutInSeconds    = 10
	maxAttempts         = 5
	initialWaitDuration = time.Second
)

type Fetcher struct {
	url        string
	httpClient *http.Client
}

type FetcherConfig struct {
	Url string
}

func New(config FetcherConfig) *Fetcher {
	return &Fetcher{url: config.Url, httpClient: &http.Client{
		Timeout: time.Second * timeoutInSeconds,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}}
}

func (f *Fetcher) Fetch(ctx context.Context) (*goquery.Document, error) {
	slog.DebugContext(ctx, "starting to fetch the web page")
	resp, err := f.fetch(ctx)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	slog.DebugContext(ctx, "fetched the web page")
	webPage, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response: %w", err)
	}

	slog.DebugContext(ctx, "read the web page as html")
	return webPage, nil
}

func (f *Fetcher) fetch(ctx context.Context) (*http.Response, error) {
	waitDuration := initialWaitDuration
	for i := 0; i < maxAttempts; i++ {
		slog.DebugContext(ctx, "attempting to fetch the web page", "attempt", i+1)
		resp, err := f.httpClient.Get(f.url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		slog.WarnContext(ctx, "failed to fetch the web page", "attempt", i+1)
		// don't sleep after the last attempt
		if i+1 >= maxAttempts {
			break
		}

		time.Sleep(waitDuration)
		waitDuration *= 2
	}

	return nil, fmt.Errorf("failed to fetch the web page after %d attempts", maxAttempts)
}
