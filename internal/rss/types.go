// Package rss fetches and parses news feeds in RSS 2.0 and Atom formats.
package rss

import (
	"encoding/xml"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// RSS is the root element of an RSS 2.0 document. Parse also converts Atom
// feeds to this form, so the rest of knews handles a single format.
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Channel contains the metadata and items of an RSS feed.
type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

// Item is a single entry in a feed.
type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"` // may contain HTML; see Summary
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Author      string `xml:"http://purl.org/dc/elements/1.1/ creator"`
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
	Feed RSS
	Err  error
}
