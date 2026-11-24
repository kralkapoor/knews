// Package rss fetches and parses news feeds in RSS 2.0 and Atom formats.
package rss

import (
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Feed is a parsed news feed. Both RSS 2.0 and Atom documents are converted
// to this form, so the rest of knews handles a single format.
type Feed struct {
	Title string
	Items []Item
}

// Item is a single entry in a feed.
type Item struct {
	Title       string
	Description string // may contain HTML; see Summary
	PubDate     string
	Link        string
	GUID        string
	Author      string
}

// pubDateLayouts are the date formats seen in feeds. RSS specifies RFC 822
// dates, but day-of-month may be one or two digits (layout "2" accepts both)
// and some feeds use ISO 8601 instead.
var pubDateLayouts = []string{
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 MST",
	time.RFC3339,
	"2006-01-02",
}

// Published parses the item's PubDate. It reports false if the date is
// missing or in an unrecognised format.
func (i Item) Published() (time.Time, bool) {
	for _, layout := range pubDateLayouts {
		if t, err := time.Parse(layout, i.PubDate); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Summary returns the item's description as plain text, with HTML markup
// removed and whitespace collapsed.
func (i Item) Summary() string {
	var b strings.Builder
	z := html.NewTokenizer(strings.NewReader(i.Description))
	for {
		switch z.Next() {
		case html.ErrorToken:
			// End of the description (or invalid HTML): return what we have.
			return strings.Join(strings.Fields(b.String()), " ")
		case html.TextToken:
			b.Write(z.Text())
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			// Keep words in separate blocks apart: "<p>a</p><p>b</p>" is "a b".
			switch name, _ := z.TagName(); string(name) {
			case "p", "br", "div", "li":
				b.WriteByte(' ')
			}
		}
	}
}

// FeedRequest names a feed to fetch.
type FeedRequest struct {
	Name string
	URL  string
}

// FeedResult is the outcome of fetching one FeedRequest. Exactly one of Feed
// and Err is meaningful.
type FeedResult struct {
	FeedRequest
	Feed Feed
	Err  error
}
