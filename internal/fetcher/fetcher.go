package fetcher

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	timeoutInSeconds = 10
	maxAttempts             = 5
	initialWaitDuration     = time.Second
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
	}}
}

func (f *Fetcher) Fetch() (*goquery.Document, error) {
	resp, err := f.fetch()
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	webPage, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response: %w", err)
	}

	return webPage, nil
}

func (f *Fetcher) fetch() (*http.Response, error) {
	waitDuration := initialWaitDuration
	for i := 0; i < maxAttempts; i++ {
		resp, err := f.httpClient.Get(f.url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		time.Sleep(waitDuration)
		waitDuration *= 2
	}

	return nil, fmt.Errorf("failed to fetch the web page after %d attempts", maxAttempts)
}
