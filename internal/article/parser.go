package article

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ArticleParser defines the interface for parsing HTML content into an Article.
// Different news sources may have different HTML structures, so each source
// should have its own parser implementation.
type ArticleParser interface {
	// Parse extracts article content from HTML bytes and the article URL.
	Parse(htmlContent []byte, articleURL string) (*Article, error)

	// CanParse returns true if this parser can handle the given URL.
	CanParse(url string) bool
}

// ParserRegistry manages multiple ArticleParser implementations and selects
// the appropriate parser based on the article URL.
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
type ArticleFetcher struct {
	httpClient *http.Client
	registry   *ParserRegistry
	userAgent  string
}

// NewArticleFetcher creates a new ArticleFetcher with the given parser registry
// and timeout duration.
func NewArticleFetcher(registry *ParserRegistry, timeout time.Duration) *ArticleFetcher {
	return &ArticleFetcher{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		registry:  registry,
		userAgent: "Knews/1.0 (NZ News Reader)",
	}
}

// Fetch retrieves and parses an article from the given URL.
func (af *ArticleFetcher) Fetch(ctx context.Context, url string) (*Article, error) {
	parser := af.registry.GetParser(url)
	if parser == nil {
		return nil, fmt.Errorf("no parser available for URL: %s", url)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("User-Agent", af.userAgent)
	request.Header.Set("Accept", "text/html")

	response, err := af.httpClient.Do(request)
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
