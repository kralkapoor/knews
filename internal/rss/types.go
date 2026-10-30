// Package rss fetches and parses RSS 2.0 feeds.
package rss

import (
	"encoding/xml"
	"time"
)

// RSS is the root element of an RSS 2.0 document.
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

// Item is a single entry in an RSS feed.
type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
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
