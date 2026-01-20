package article

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// RNZParser implements ArticleParser for Radio New Zealand articles.
type RNZParser struct{}

// NewRNZParser creates a new RNZParser instance.
func NewRNZParser() *RNZParser {
	return &RNZParser{}
}

// CanParse returns true if the URL is from the RNZ website.
func (p *RNZParser) CanParse(url string) bool {
	return strings.Contains(url, "rnz.co.nz")
}

// Parse extracts article content from RNZ HTML.
func (p *RNZParser) Parse(htmlContent []byte, articleURL string) (*Article, error) {
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	article := &Article{
		URL:    articleURL,
		Source: "RNZ",
	}

	article.Title = p.extractTitle(document)
	article.Author = p.extractAuthor(document)
	article.PublishedAt = p.extractPublishedDate(document, htmlContent)
	article.Content = p.extractContent(document)
	article.Tags = p.extractTags(document)
	article.Images = p.extractImages(document)

	return article, nil
}

// extractTitle extracts the article title from the document.
func (p *RNZParser) extractTitle(document *goquery.Document) string {
	// Try multiple selectors for title
	titleSelectors := []string{
		"h1.article__title",
		"h1.c-article__title",
		"article h1",
		"h1",
	}

	for _, selector := range titleSelectors {
		title := strings.TrimSpace(document.Find(selector).First().Text())
		if title != "" {
			return title
		}
	}

	return ""
}

// extractAuthor extracts the author information from the document.
func (p *RNZParser) extractAuthor(document *goquery.Document) Author {
	// Try multiple selectors for author
	authorSelectors := []string{
		".article__byline a",
		".c-article__byline a",
		".byline a",
		".author a",
	}

	for _, selector := range authorSelectors {
		authorElement := document.Find(selector).First()
		name := strings.TrimSpace(authorElement.Text())
		if name != "" {
			url, _ := authorElement.Attr("href")
			return Author{Name: name, URL: url}
		}
	}

	// Try without link
	textSelectors := []string{
		".article__byline",
		".c-article__byline",
		".byline",
	}

	for _, selector := range textSelectors {
		name := strings.TrimSpace(document.Find(selector).First().Text())
		name = strings.TrimPrefix(name, "By ")
		if name != "" {
			return Author{Name: name}
		}
	}

	return Author{}
}

// extractPublishedDate extracts the publication date from JSON-LD or other sources.
func (p *RNZParser) extractPublishedDate(document *goquery.Document, htmlContent []byte) time.Time {
	var publishedDate time.Time

	// Try to extract from JSON-LD schema
	document.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		if !publishedDate.IsZero() {
			return
		}
		jsonText := s.Text()
		var schemaData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonText), &schemaData); err == nil {
			if datePublished, ok := schemaData["datePublished"].(string); ok {
				if parsed, err := time.Parse(time.RFC3339, datePublished); err == nil {
					publishedDate = parsed
					return
				}
				if parsed, err := time.Parse("2006-01-02T15:04:05", datePublished); err == nil {
					publishedDate = parsed
					return
				}
			}
		}
	})

	if !publishedDate.IsZero() {
		return publishedDate
	}

	// Try to extract from time element
	timeElement := document.Find("time[datetime]").First()
	if datetime, exists := timeElement.Attr("datetime"); exists {
		if parsed, err := time.Parse(time.RFC3339, datetime); err == nil {
			return parsed
		}
		if parsed, err := time.Parse("2006-01-02", datetime); err == nil {
			return parsed
		}
	}

	// Try to extract from meta tag
	metaDate := document.Find("meta[property='article:published_time']").First()
	if content, exists := metaDate.Attr("content"); exists {
		if parsed, err := time.Parse(time.RFC3339, content); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

// extractContent extracts the main article content blocks.
func (p *RNZParser) extractContent(document *goquery.Document) []ContentBlock {
	var contentBlocks []ContentBlock

	// Try multiple selectors for article body
	bodySelectors := []string{
		".article__body",
		".c-article__body",
		".article-body",
		"article .content",
		"article",
	}

	var articleBody *goquery.Selection
	for _, selector := range bodySelectors {
		selection := document.Find(selector).First()
		if selection.Length() > 0 {
			articleBody = selection
			break
		}
	}

	if articleBody == nil {
		return contentBlocks
	}

	// Extract paragraphs, quotes, and headings
	articleBody.Find("p, blockquote, h2, h3, h4").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		tagName := goquery.NodeName(s)
		var block ContentBlock

		switch tagName {
		case "blockquote":
			block = ContentBlock{
				Type:    BlockTypeQuote,
				Content: text,
			}
		case "h2":
			block = ContentBlock{
				Type:    BlockTypeHeading,
				Content: text,
				Level:   2,
			}
		case "h3":
			block = ContentBlock{
				Type:    BlockTypeHeading,
				Content: text,
				Level:   3,
			}
		case "h4":
			block = ContentBlock{
				Type:    BlockTypeHeading,
				Content: text,
				Level:   4,
			}
		default:
			block = ContentBlock{
				Type:    BlockTypeParagraph,
				Content: text,
			}
		}

		contentBlocks = append(contentBlocks, block)
	})

	return contentBlocks
}

// extractTags extracts article tags from tag links.
func (p *RNZParser) extractTags(document *goquery.Document) []string {
	var tags []string
	seenTags := make(map[string]bool)

	// Look for tag links
	tagSelectors := []string{
		"a[href*='/tags/']",
		".article__tags a",
		".tags a",
	}

	for _, selector := range tagSelectors {
		document.Find(selector).Each(func(i int, s *goquery.Selection) {
			tag := strings.TrimSpace(s.Text())
			if tag != "" && !seenTags[tag] {
				tags = append(tags, tag)
				seenTags[tag] = true
			}
		})
	}

	return tags
}

// extractImages extracts images with their captions and alt text.
func (p *RNZParser) extractImages(document *goquery.Document) []Image {
	var images []Image

	// Look for images in figure elements
	document.Find("figure").Each(func(i int, figure *goquery.Selection) {
		img := figure.Find("img").First()
		if img.Length() == 0 {
			return
		}

		src, _ := img.Attr("src")
		if src == "" {
			src, _ = img.Attr("data-src")
		}
		if src == "" {
			return
		}

		alt, _ := img.Attr("alt")
		caption := strings.TrimSpace(figure.Find("figcaption").Text())

		images = append(images, Image{
			URL:     src,
			Caption: caption,
			Alt:     alt,
		})
	})

	// Also look for standalone images in article body
	document.Find(".article__body img, .c-article__body img").Each(func(i int, img *goquery.Selection) {
		// Skip if already captured in a figure
		if img.ParentsFiltered("figure").Length() > 0 {
			return
		}

		src, _ := img.Attr("src")
		if src == "" {
			src, _ = img.Attr("data-src")
		}
		if src == "" {
			return
		}

		alt, _ := img.Attr("alt")

		images = append(images, Image{
			URL: src,
			Alt: alt,
		})
	})

	return images
}
