package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type FeedClient struct{}

func NewFeedClient() *FeedClient {
	return &FeedClient{}
}

func (fc *FeedClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", response.StatusCode, url)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return bodyBytes, nil
}

// FetchAndParse fetches the RSS feed from the URL and parses it into an Rss struct.
func (fc *FeedClient) FetchAndParse(ctx context.Context, url string) (Rss, error) {
	bodyBytes, err := fc.Fetch(ctx, url)
	if err != nil {
		return Rss{}, err
	}

	rss, err := ParseXmlResponse(bodyBytes)
	if err != nil {
		return Rss{}, fmt.Errorf("failed to parse RSS from %s: %w", url, err)
	}

	return rss, nil
}
