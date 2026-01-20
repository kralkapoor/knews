package article

import (
	"bytes"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// RNZParser implements ArticleParser for Radio New Zealand articles.
type RNZParser struct{}

// NewRNZParser creates a new RNZParser instance.
func NewRNZParser() *RNZParser {
	return &RNZParser{}
}

// CanParse returns true if the URL is from the RNZ website.
func (p *RNZParser) CanParse(articleURL string) bool {
	return strings.Contains(articleURL, "rnz.co.nz")
}

// Parse extracts article content from RNZ HTML using go-readability.
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
