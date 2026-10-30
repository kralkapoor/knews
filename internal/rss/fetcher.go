package rss

import (
	"context"
	"sync"
)

const defaultMaxWorkers = 5

// Fetcher fetches many feeds concurrently with a bounded pool of workers.
type Fetcher struct {
	client     *Client
	maxWorkers int
}

// NewFetcher returns a Fetcher that runs at most maxWorkers requests at once.
// A maxWorkers of zero or less selects a default of 5.
func NewFetcher(client *Client, maxWorkers int) *Fetcher {
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}
	return &Fetcher{client: client, maxWorkers: maxWorkers}
}

// FetchAll fetches every request and returns one result per request, in the
// same order. Requests not yet started when ctx is done fail with ctx.Err().
func (f *Fetcher) FetchAll(ctx context.Context, requests []FeedRequest) []FeedResult {
	if len(requests) == 0 {
		return nil
	}

	results := make([]FeedResult, len(requests))
	pending := make(chan int, len(requests))
	for i, req := range requests {
		results[i].FeedRequest = req
		pending <- i
	}
	close(pending)

	// Each worker writes only to the result slots it takes from pending, so
	// no further synchronisation is needed.
	var wg sync.WaitGroup
	for range min(f.maxWorkers, len(requests)) {
		wg.Go(func() {
			for i := range pending {
				if err := ctx.Err(); err != nil {
					results[i].Err = err
					continue
				}
				results[i].Feed, results[i].Err = f.client.Fetch(ctx, requests[i].URL)
			}
		})
	}
	wg.Wait()

	return results
}
