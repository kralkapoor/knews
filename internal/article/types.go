// Package article fetches news article pages and extracts their text.
package article

import "time"

// Article is a news article extracted from its web page.
type Article struct {
	URL         string
	Title       string
	Author      string
	Content     string
	PublishedAt time.Time // zero if the page has no publication date
	Source      string
}
