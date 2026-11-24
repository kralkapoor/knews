// Package httpx holds the HTTP plumbing shared by the feed and article fetchers.
package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// UserAgent identifies knews to news sites. Some CDNs throttle Go's default
	// "Go-http-client/1.1" agent.
	UserAgent = "knews/1.0 (+https://github.com/kralkapoor/knews)"

	// DefaultTimeout bounds a whole request, including reading the body.
	DefaultTimeout = 30 * time.Second

	// MaxBodyBytes caps how much of a response body Get reads (10 MB).
	MaxBodyBytes = 10 * 1024 * 1024
)

// NewClient returns an HTTP client with DefaultTimeout.
func NewClient() *http.Client {
	return &http.Client{Timeout: DefaultTimeout}
}

// Get fetches url and returns its body. Responses other than 200 OK and bodies
// larger than MaxBodyBytes are errors.
func Get(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", url, err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Drain a little of the body (64 KB) so the connection can be reused.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		return nil, fmt.Errorf("get %s: unexpected status %s", url, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("get %s: read body: %w", url, err)
	}
	if len(body) > MaxBodyBytes {
		return nil, fmt.Errorf("get %s: body exceeds %d bytes", url, MaxBodyBytes)
	}
	return body, nil
}
