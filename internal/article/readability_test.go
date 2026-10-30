package article

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestReadabilityParserCanParse(t *testing.T) {
	p := ReadabilityParser{Hosts: []string{"rnz.co.nz"}}
	tests := []struct {
		url  string
		want bool
	}{
		{"https://www.rnz.co.nz/news/political/584370/story", true},
		{"https://rnz.co.nz/news", true},
		{"http://WWW.RNZ.CO.NZ/news", true},
		{"https://www.stuff.co.nz/nz-news/1/story", false},
		{"https://notrnz.co.nz/news", false},
		{"https://evil.example/?u=rnz.co.nz", false},
		{"https://rnz.co.nz.evil.example/", false},
		{"", false},
		{"::not a url", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := p.CanParse(tt.url); got != tt.want {
				t.Errorf("CanParse(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestReadabilityParserParse(t *testing.T) {
	html, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}
	p := ReadabilityParser{Source: "RNZ", Hosts: []string{"rnz.co.nz"}}

	const url = "https://www.rnz.co.nz/news/environment/1/kakapo"
	art, err := p.Parse(html, url)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if art.URL != url {
		t.Errorf("URL = %q, want %q", art.URL, url)
	}
	if art.Source != "RNZ" {
		t.Errorf("Source = %q, want RNZ", art.Source)
	}
	if !strings.Contains(art.Title, "Kākāpō breeding season smashes records") {
		t.Errorf("Title = %q", art.Title)
	}
	if !strings.Contains(art.Author, "Jane Smith") {
		t.Errorf("Author = %q, want it to contain Jane Smith", art.Author)
	}
	if got := art.PublishedAt.Format("2006-01-02"); got != "2026-03-04" {
		t.Errorf("PublishedAt = %v, want 2026-03-04", art.PublishedAt)
	}
	if !strings.Contains(art.Content, "82 chicks had survived") {
		t.Errorf("Content is missing the article body:\n%s", art.Content)
	}
	if strings.Contains(art.Content, "Copyright RNZ") {
		t.Errorf("Content includes page chrome:\n%s", art.Content)
	}
}

func TestReadabilityParserParseClientRenderedPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html><html><head><title>Story</title></head>` +
		`<body><div id="root"></div><script src="/main.js"></script></body></html>`)

	_, err := ReadabilityParser{}.Parse(html, "https://www.stuff.co.nz/nz-news/1/story")
	if !errors.Is(err, ErrNoContent) {
		t.Fatalf("Parse() error = %v, want ErrNoContent", err)
	}
}
