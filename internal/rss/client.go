package rss

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kralkapoor/knews/internal/httpx"
)

// Client fetches and parses individual feeds.
type Client struct {
	http *http.Client
}

// NewClient returns a Client that sends requests with hc, or with
// httpx.NewClient() if hc is nil.
func NewClient(hc *http.Client) *Client {
	if hc == nil {
		hc = httpx.NewClient()
	}
	return &Client{http: hc}
}

// Fetch downloads the feed at url and parses it.
func (c *Client) Fetch(ctx context.Context, url string) (RSS, error) {
	body, err := httpx.Get(ctx, c.http, url)
	if err != nil {
		return RSS{}, err
	}

	feed, err := Parse(body)
	if err != nil {
		return RSS{}, fmt.Errorf("parse feed %s: %w", url, err)
	}
	return feed, nil
}
