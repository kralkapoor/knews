package article

import (
	"bytes"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// RNZParser implements the ArticleParser interface for Radio New Zealand articles.
//
// Go knows this implements ArticleParser because it has the required methods:
//   - CanParse(url string) bool
//   - Parse(htmlContent []byte, articleURL string) (*Article, error)
//
// No explicit declaration needed - this is "duck typing" (if it walks like a duck...).
// The compiler will catch any mistakes: if you try to register an RNZParser with a
// registry that expects ArticleParser, and RNZParser is missing a method, you'll
// get a compile error.
type RNZParser struct{}

// NewRNZParser creates a new RNZParser instance.
func NewRNZParser() *RNZParser {
	return &RNZParser{}
}

// CanParse returns true if the URL is from the RNZ website.
// This method is part of the ArticleParser interface.
func (p *RNZParser) CanParse(articleURL string) bool {
	return strings.Contains(articleURL, "rnz.co.nz")
}

// Parse extracts article content from RNZ HTML using go-readability.
// This method is part of the ArticleParser interface.
//
// The go-readability library does the heavy lifting of extracting the main
// content from the HTML, similar to Firefox's Reader View.
func (p *RNZParser) Parse(htmlContent []byte, articleURL string) (*Article, error) {
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return nil, err
	}

	doc, err := readability.FromReader(bytes.NewReader(htmlContent), parsedURL)
	if err != nil {
		return nil, err
	}

	publishedAt, _ := doc.PublishedTime()

	var contentBuffer bytes.Buffer
	doc.RenderText(&contentBuffer)

	return &Article{
		URL:         articleURL,
		Title:       doc.Title(),
		Author:      doc.Byline(),
		Content:     contentBuffer.String(),
		PublishedAt: publishedAt,
		Source:      "RNZ",
	}, nil
}
