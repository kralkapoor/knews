// Package article fetches news article pages and extracts their text.
package article

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/kralkapoor/knews/internal/httpx"
)

// ErrUnsupportedHost is returned by Fetcher.Fetch for URLs whose host matches
// none of the fetcher's rules.
var ErrUnsupportedHost = errors.New("unsupported article host")

// ErrNoContent is returned when a page contains no extractable article text,
// for example because the site renders its articles with JavaScript.
var ErrNoContent = errors.New("no article content found")

// Article is a news article extracted from its web page.
type Article struct {
	URL         string
	Title       string
	Author      string
	Content     string
	PublishedAt time.Time // zero if the page has no publication date
	Source      string
}

// HostRule says which websites belong to a news source.
type HostRule struct {
	// Source labels extracted articles, e.g. "RNZ".
	Source string

	// Hosts are the domains the source's articles are served from.
	// Subdomains match too, so "rnz.co.nz" covers "www.rnz.co.nz".
	Hosts []string

	// Extract pulls the article out of a fetched page. Nil means the
	// package's Extract (go-readability), which suits most server-rendered
	// news sites. Set it for a site readability handles badly.
	Extract func(html []byte, articleURL, source string) (*Article, error)
}

// Fetcher downloads article pages and extracts their text.
type Fetcher struct {
	httpClient *http.Client
	rules      []HostRule
}

// NewFetcher returns a Fetcher that only fetches URLs matching one of rules.
// It sends requests with httpClient, or with httpx.NewClient() if httpClient
// is nil.
func NewFetcher(httpClient *http.Client, rules []HostRule) *Fetcher {
	if httpClient == nil {
		httpClient = httpx.NewClient()
	}
	return &Fetcher{httpClient: httpClient, rules: rules}
}

// Fetch downloads the article at articleURL and extracts it with the matching
// rule's Extract. It returns an error wrapping ErrUnsupportedHost, without
// making a request, if no rule matches the URL's host.
func (f *Fetcher) Fetch(ctx context.Context, articleURL string) (*Article, error) {
	rule, ok := f.ruleFor(articleURL)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedHost, articleURL)
	}

	body, err := httpx.Get(ctx, f.httpClient, articleURL)
	if err != nil {
		return nil, err
	}

	extract := rule.Extract
	if extract == nil {
		extract = Extract
	}
	return extract(body, articleURL, rule.Source)
}

// ruleFor returns the first rule whose hosts match articleURL.
func (f *Fetcher) ruleFor(articleURL string) (HostRule, bool) {
	for _, rule := range f.rules {
		if matchesHost(articleURL, rule.Hosts) {
			return rule, true
		}
	}
	return HostRule{}, false
}

// matchesHost reports whether articleURL's host is one of hosts or a
// subdomain of one. It compares the parsed host, so
// "https://evil.example/?u=rnz.co.nz" doesn't match "rnz.co.nz".
func matchesHost(articleURL string, hosts []string) bool {
	u, err := url.Parse(articleURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, h := range hosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}
	return false
}

// Extract pulls the article out of html, which was fetched from articleURL,
// using go-readability: a port of the algorithm behind Firefox's Reader View.
// It works for most server-rendered news sites.
func Extract(html []byte, articleURL string, source string) (*Article, error) {
	pageURL, err := url.Parse(articleURL)
	if err != nil {
		return nil, fmt.Errorf("parse article URL: %w", err)
	}

	doc, err := readability.FromReader(bytes.NewReader(html), pageURL)
	if err != nil {
		return nil, fmt.Errorf("extract %s: %w", articleURL, err)
	}
	if doc.Node == nil {
		return nil, fmt.Errorf("%w at %s", ErrNoContent, articleURL)
	}

	var content strings.Builder
	if err := doc.RenderText(&content); err != nil {
		return nil, fmt.Errorf("render %s: %w", articleURL, err)
	}
	text := strings.TrimSpace(content.String())
	if text == "" {
		return nil, fmt.Errorf("%w at %s", ErrNoContent, articleURL)
	}

	// A missing or malformed date leaves publishedAt zero, which the UI hides.
	publishedAt, _ := doc.PublishedTime()

	return &Article{
		URL:         articleURL,
		Title:       doc.Title(),
		Author:      doc.Byline(),
		Content:     text,
		PublishedAt: publishedAt,
		Source:      source,
	}, nil
}
