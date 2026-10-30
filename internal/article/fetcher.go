package article

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/kralkapoor/knews/internal/httpx"
)

// ErrNoParser is returned by Fetcher.Fetch for URLs no registered parser handles.
var ErrNoParser = errors.New("no parser for URL")

// Fetcher downloads article pages and extracts them with a registered Parser.
type Fetcher struct {
	registry *Registry
	http     *http.Client
}

// NewFetcher returns a Fetcher that uses parsers from registry and sends
// requests with hc, or with httpx.NewClient() if hc is nil.
func NewFetcher(registry *Registry, hc *http.Client) *Fetcher {
	if hc == nil {
		hc = httpx.NewClient()
	}
	return &Fetcher{registry: registry, http: hc}
}

// Fetch downloads and parses the article at url. It returns an error wrapping
// ErrNoParser, without making a request, if no parser handles url.
func (f *Fetcher) Fetch(ctx context.Context, url string) (*Article, error) {
	parser := f.registry.parserFor(url)
	if parser == nil {
		return nil, fmt.Errorf("%w: %s", ErrNoParser, url)
	}

	body, err := httpx.Get(ctx, f.http, url)
	if err != nil {
		return nil, err
	}
	return parser.Parse(body, url)
}
