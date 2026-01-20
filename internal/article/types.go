package article

import "time"

// BlockType represents the type of content block in an article.
type BlockType int

const (
	BlockTypeParagraph BlockType = iota
	BlockTypeQuote
	BlockTypeHeading
	BlockTypeList
)

// String returns a human-readable representation of the BlockType.
func (bt BlockType) String() string {
	switch bt {
	case BlockTypeParagraph:
		return "paragraph"
	case BlockTypeQuote:
		return "quote"
	case BlockTypeHeading:
		return "heading"
	case BlockTypeList:
		return "list"
	default:
		return "unknown"
	}
}

// ContentBlock represents a single block of content within an article,
// such as a paragraph, quote, or heading.
type ContentBlock struct {
	Type    BlockType
	Content string
	Level   int // For headings, the heading level (1-6)
}

// Image represents an image within an article with its associated metadata.
type Image struct {
	URL     string
	Caption string
	Alt     string
}

// Author represents the author of an article.
type Author struct {
	Name string
	URL  string
}

// Article represents a fully parsed news article with all its content
// and metadata extracted from the source HTML.
type Article struct {
	URL         string
	Title       string
	Content     []ContentBlock
	Author      Author
	PublishedAt time.Time
	Tags        []string
	Images      []Image
	Source      string // e.g., "RNZ", "NZ Herald"
}

// PlainTextContent returns the article content as plain text, concatenating
// all content blocks with appropriate formatting.
func (a *Article) PlainTextContent() string {
	var result string
	for i, block := range a.Content {
		if i > 0 {
			result += "\n\n"
		}
		switch block.Type {
		case BlockTypeQuote:
			result += "> " + block.Content
		case BlockTypeHeading:
			result += block.Content
		default:
			result += block.Content
		}
	}
	return result
}
