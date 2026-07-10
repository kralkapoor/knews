package rss

import "time"

// Rss represents the root element of an RSS feed.
type Rss struct {
	Channel Channel `xml:"channel"`
}

// Channel contains the metadata and items of an RSS feed.
type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Copyright   string `xml:"copyright"`
	Items       []Item `xml:"item"`
}

// Item represents a single entry in an RSS feed.
type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
}

// FeedResult encapsulates the result of fetching a single RSS feed, including
// success data and any error that occurred during the fetch operation.
type FeedResult struct {
	Feed      Rss
	SourceURL string
	FeedName  string
	FetchedAt time.Time
	Error     error
}

// FeedRequest represents a request to fetch a feed with its metadata.
type FeedRequest struct {
	URL  string
	Name string
}
