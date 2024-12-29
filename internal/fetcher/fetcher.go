package fetcher

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher struct {
	url        string
	httpClient *http.Client
}

func New(url string) *Fetcher {
	return &Fetcher{url: url, httpClient: &http.Client{
		Timeout: time.Second * 10,
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
