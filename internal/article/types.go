package article

import "time"

// Article represents a fully parsed news article with all its content
// and metadata extracted from the source HTML.
type Article struct {
	URL         string
	Title       string
	Author      string
	Content     string
	PublishedAt time.Time
	Source      string
}
