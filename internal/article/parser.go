package article

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// ArticleParser defines the interface for parsing HTML content into an Article.
//
// In Go, interfaces are implemented IMPLICITLY - there's no "implements" keyword.
// Any type that has these exact method signatures automatically satisfies this
// interface. See RNZParser in rnz_parser.go for an example.
type ArticleParser interface {
	// Parse extracts article content from HTML bytes and the article URL.
	Parse(htmlContent []byte, articleURL string) (*Article, error)

	// CanParse returns true if this parser can handle the given URL.
	CanParse(url string) bool
}

// ParserRegistry holds multiple parsers and finds the right one for a URL.
//
// This is the "registry pattern" - instead of hardcoding which parser to use,
// we loop through registered parsers and ask each one "can you handle this URL?"
// This makes it easy to add support for new news sources later.
type ParserRegistry struct {
	parsers []ArticleParser
}

// NewParserRegistry creates a new empty ParserRegistry.
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make([]ArticleParser, 0),
	}
}

// Register adds a parser to the registry.
func (pr *ParserRegistry) Register(parser ArticleParser) {
	pr.parsers = append(pr.parsers, parser)
}

// GetParser returns the first parser that can handle the given URL,
// or nil if no suitable parser is found.
func (pr *ParserRegistry) GetParser(url string) ArticleParser {
	for _, parser := range pr.parsers {
		if parser.CanParse(url) {
			return parser
		}
	}
	return nil
}

// Parse finds an appropriate parser for the URL and parses the HTML content.
func (pr *ParserRegistry) Parse(htmlContent []byte, articleURL string) (*Article, error) {
	parser := pr.GetParser(articleURL)
	if parser == nil {
		return nil, fmt.Errorf("no parser available for URL: %s", articleURL)
	}
	return parser.Parse(htmlContent, articleURL)
}

// ArticleFetcher handles fetching and parsing articles from URLs.
// It uses the registry to find the appropriate parser for each URL.
type ArticleFetcher struct {
	registry *ParserRegistry
}

// NewArticleFetcher creates a new ArticleFetcher with the given parser registry.
func NewArticleFetcher(registry *ParserRegistry) *ArticleFetcher {
	return &ArticleFetcher{registry: registry}
}

// Fetch retrieves and parses an article from the given URL.
//
// The context parameter (ctx) controls timeouts and cancellation. When the caller
// creates a context with context.WithTimeout(), the HTTP request will automatically
// be cancelled if it takes too long. This is why we don't need a custom http.Client
// with its own timeout - the context handles it for us.
func (af *ArticleFetcher) Fetch(ctx context.Context, url string) (*Article, error) {
	parser := af.registry.GetParser(url)
	if parser == nil {
		return nil, fmt.Errorf("no parser available for URL: %s", url)
	}

	// NewRequestWithContext creates an HTTP request that respects context cancellation.
	// If the context times out or is cancelled, the request will be aborted.
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// The %w verb wraps the original error, preserving the error chain.
		// This lets callers use errors.Is() or errors.As() to check the underlying error.
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// http.DefaultClient is Go's built-in HTTP client. It respects context
	// cancellation, so the timeout set by the caller's context will work.
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", response.StatusCode)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return parser.Parse(bodyBytes, url)
}
