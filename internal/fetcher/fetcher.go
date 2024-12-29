package fetcher

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const defaultTimeoutInSeconds = 10

type Fetcher struct {
	url        string
	httpClient *http.Client
}

type FetcherConfig struct {
	Url string
}

func New(config FetcherConfig) *Fetcher {
	return &Fetcher{url: config.Url, httpClient: &http.Client{
		Timeout: time.Second * defaultTimeoutInSeconds,
	}}
}

// TODO: retry logic
func (f *Fetcher) Fetch() (*goquery.Document, error) {
	resp, err := f.httpClient.Get(f.url)
	if err != nil {
		return nil, fmt.Errorf("error fetching the web page: %w", err)
	}

	defer resp.Body.Close()
	webPage, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response: %w", err)
	}

	return webPage, nil
}
