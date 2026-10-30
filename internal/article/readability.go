package article

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// ErrNoContent is returned when a page contains no extractable article text,
// for example because the site renders its articles with JavaScript.
var ErrNoContent = errors.New("no article content found")

// ReadabilityParser extracts articles with go-readability, a port of the
// algorithm behind Firefox's Reader View. It works for most server-rendered
// news sites, so a new provider usually needs only a configured instance.
type ReadabilityParser struct {
	// Source labels extracted articles, e.g. "RNZ".
	Source string

	// Hosts are the domains this parser handles. Subdomains match too, so
	// "rnz.co.nz" covers "www.rnz.co.nz".
	Hosts []string
}

var _ Parser = ReadabilityParser{}

// CanParse reports whether articleURL's host is one of p.Hosts or a subdomain
// of one.
func (p ReadabilityParser) CanParse(articleURL string) bool {
	u, err := url.Parse(articleURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, h := range p.Hosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}
	return false
}

// Parse extracts the article from html.
func (p ReadabilityParser) Parse(html []byte, articleURL string) (*Article, error) {
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
		Source:      p.Source,
	}, nil
}
