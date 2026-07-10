package rss

import (
	"context"
	"sync"
	"time"
)

const defaultMaxWorkers = 5

type FeedFetcher struct {
	client     *FeedClient
	maxWorkers int
}

func NewFeedFetcher(client *FeedClient, maxWorkers int) *FeedFetcher {
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}

	return &FeedFetcher{
		client:     client,
		maxWorkers: maxWorkers,
	}
}

func (ff *FeedFetcher) FetchAll(ctx context.Context, requests []FeedRequest) []FeedResult {
	if len(requests) == 0 {
		return nil
	}

	results := make([]FeedResult, len(requests))

	requestChannel := make(chan int, len(requests))

	var waitGroup sync.WaitGroup

	workerCount := min(ff.maxWorkers, len(requests))

	for range workerCount {
		waitGroup.Add(1)
		go ff.worker(ctx, requests, results, requestChannel, &waitGroup)
	}

	for requestIndex := range requests {
		requestChannel <- requestIndex
	}
	close(requestChannel)

	waitGroup.Wait()

	return results
}

// worker processes feed requests from the channel and stores results.
func (ff *FeedFetcher) worker(
	ctx context.Context,
	requests []FeedRequest,
	results []FeedResult,
	requestChannel <-chan int,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	for requestIndex := range requestChannel {
		select {
		case <-ctx.Done():
			results[requestIndex] = FeedResult{
				SourceURL: requests[requestIndex].URL,
				FeedName:  requests[requestIndex].Name,
				FetchedAt: time.Now(),
				Error:     ctx.Err(),
			}
			return
		default:
			results[requestIndex] = ff.fetchSingle(ctx, requests[requestIndex])
		}
	}
}

// fetchSingle fetches and parses a single feed request.
func (ff *FeedFetcher) fetchSingle(ctx context.Context, request FeedRequest) FeedResult {
	fetchedAt := time.Now()

	feed, err := ff.client.FetchAndParse(ctx, request.URL)

	return FeedResult{
		Feed:      feed,
		SourceURL: request.URL,
		FeedName:  request.Name,
		FetchedAt: fetchedAt,
		Error:     err,
	}
}

// FetchAllURLs is a convenience method that fetches feeds from a slice of URLs
// using empty names. Useful when feed names are not needed.
func (ff *FeedFetcher) FetchAllURLs(ctx context.Context, urls []string) []FeedResult {
	requests := make([]FeedRequest, len(urls))
	for i, url := range urls {
		requests[i] = FeedRequest{URL: url}
	}
	return ff.FetchAll(ctx, requests)
}

// FetchAllFromMap fetches feeds from a map of name to URL.
func (ff *FeedFetcher) FetchAllFromMap(ctx context.Context, feeds map[string]string) []FeedResult {
	requests := make([]FeedRequest, 0, len(feeds))
	for name, url := range feeds {
		requests = append(requests, FeedRequest{URL: url, Name: name})
	}
	return ff.FetchAll(ctx, requests)
}
