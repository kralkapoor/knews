package rss

import (
	"testing"
)

func TestParseXmlResponse(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <description>A test feed</description>
    <link>https://example.com</link>
    <item>
      <title>Test Article</title>
      <description>Test description</description>
      <link>https://example.com/article</link>
      <pubDate>Mon, 20 Jan 2026 10:00:00 +0000</pubDate>
      <guid>article-1</guid>
    </item>
  </channel>
</rss>`)

	rss, err := ParseXmlResponse(xmlData)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if rss.Channel.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", rss.Channel.Title)
	}

	if len(rss.Channel.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(rss.Channel.Items))
	}

	item := rss.Channel.Items[0]
	if item.Title != "Test Article" {
		t.Errorf("Expected item title 'Test Article', got '%s'", item.Title)
	}

	if item.GUID != "article-1" {
		t.Errorf("Expected GUID 'article-1', got '%s'", item.GUID)
	}
}

func TestParseXmlResponseInvalidXml(t *testing.T) {
	invalidXml := []byte(`not valid xml`)

	_, err := ParseXmlResponse(invalidXml)
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}
