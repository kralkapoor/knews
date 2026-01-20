package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "Knews/1.0 (NZ News Reader)"
)

// FeedClient provides HTTP client functionality for fetching RSS feeds with
// proper timeout handling, context support, and error management.
type FeedClient struct {
	httpClient *http.Client
	userAgent  string
}

// NewFeedClient creates a new FeedClient with the specified timeout duration.
// If timeout is zero or negative, the default timeout of 10 seconds is used.
func NewFeedClient(timeout time.Duration) *FeedClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &FeedClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: defaultUserAgent,
	}
}

// NewFeedClientWithUserAgent creates a new FeedClient with custom timeout and user agent.
func NewFeedClientWithUserAgent(timeout time.Duration, userAgent string) *FeedClient {
	client := NewFeedClient(timeout)
	if userAgent != "" {
		client.userAgent = userAgent
	}
	return client
}

// Fetch retrieves the content from the specified URL with context support for
// cancellation and timeouts. Returns the response body as bytes or an error.
func (fc *FeedClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	request.Header.Set("User-Agent", fc.userAgent)
	request.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	response, err := fc.httpClient.Do(request)
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
