package rss

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kralkapoor/knews/internal/httpx"
)

const defaultMaxWorkers = 5

// Fetcher downloads feeds, one at a time with Fetch or many concurrently
// with FetchAll.
type Fetcher struct {
	httpClient *http.Client

	// MaxWorkers is the most feeds FetchAll downloads at once. Zero or less
	// means defaultMaxWorkers.
	MaxWorkers int
}

// NewFetcher returns a Fetcher that sends requests with httpClient, or with
// httpx.NewClient() if httpClient is nil.
func NewFetcher(httpClient *http.Client) *Fetcher {
	if httpClient == nil {
		httpClient = httpx.NewClient()
	}
	return &Fetcher{httpClient: httpClient, MaxWorkers: defaultMaxWorkers}
}

// Fetch downloads the feed at url and parses it.
func (f *Fetcher) Fetch(ctx context.Context, url string) (Feed, error) {
	body, err := httpx.Get(ctx, f.httpClient, url)
	if err != nil {
		return Feed{}, err
	}

	feed, err := Parse(body)
	if err != nil {
		return Feed{}, fmt.Errorf("parse feed %s: %w", url, err)
	}
	return feed, nil
}

// FetchAll fetches every request and returns one result per request, in the
// same order. Requests not yet started when ctx is done fail with ctx.Err().
func (f *Fetcher) FetchAll(ctx context.Context, requests []FeedRequest) []FeedResult {
	if len(requests) == 0 {
		return nil
	}

	// Fill in every result's request up front, so each result says which feed
	// it is for even if the fetch fails. Put each request's index on a channel
	// for the workers to pick up.
	results := make([]FeedResult, len(requests))
	pending := make(chan int, len(requests))
	for i, req := range requests {
		results[i].FeedRequest = req
		pending <- i
	}
	close(pending)

	workers := f.MaxWorkers
	if workers <= 0 {
		workers = defaultMaxWorkers
	}

	// Each worker writes only to the result slots it takes from pending, so
	// no locking is needed. Never start more workers than there are requests.
	var wg sync.WaitGroup
	for range min(workers, len(requests)) {
		wg.Go(func() {
			for i := range pending {
				if err := ctx.Err(); err != nil {
					// Keep draining the channel so every result is filled in.
					results[i].Err = err
					continue
				}
				results[i].Feed, results[i].Err = f.Fetch(ctx, requests[i].URL)
			}
		})
	}
	wg.Wait()

	return results
}
