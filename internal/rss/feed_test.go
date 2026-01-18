package rss

import (
	"testing"
)

func TestStuff(t *testing.T) {
	feeds := listFeeds()
	feed := feeds[0]

	getFeed(feed)
}
