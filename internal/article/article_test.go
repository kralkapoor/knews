package article

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestMatchesHost(t *testing.T) {
	hosts := []string{"rnz.co.nz"}
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
			if got := matchesHost(tt.url, hosts); got != tt.want {
				t.Errorf("matchesHost(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	html, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}

	const url = "https://www.rnz.co.nz/news/environment/1/kakapo"
	art, err := Extract(html, url, "RNZ")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
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

func TestExtractClientRenderedPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html><html><head><title>Story</title></head>` +
		`<body><div id="root"></div><script src="/main.js"></script></body></html>`)

	_, err := Extract(html, "https://www.stuff.co.nz/nz-news/1/story", "Stuff")
	if !errors.Is(err, ErrNoContent) {
		t.Fatalf("Extract() error = %v, want ErrNoContent", err)
	}
}

func TestFetcherFetch(t *testing.T) {
	html, err := os.ReadFile("testdata/article.html")
	if err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write(html)
	}))
	defer srv.Close()

	rules := []HostRule{{Source: "Test", Hosts: []string{"127.0.0.1"}}}
	f := NewFetcher(srv.Client(), rules)

	t.Run("success", func(t *testing.T) {
		art, err := f.Fetch(context.Background(), srv.URL+"/story")
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if art.Source != "Test" || !strings.Contains(art.Content, "82 chicks") {
			t.Errorf("Fetch() = %+v", art)
		}
	})

	t.Run("http error", func(t *testing.T) {
		_, err := f.Fetch(context.Background(), srv.URL+"/missing")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("Fetch() error = %v, want 404 error", err)
		}
	})

	t.Run("unsupported host", func(t *testing.T) {
		before := hits.Load()
		_, err := f.Fetch(context.Background(), "https://unknown.example/story")
		if !errors.Is(err, ErrUnsupportedHost) {
			t.Fatalf("Fetch() error = %v, want ErrUnsupportedHost", err)
		}
		if hits.Load() != before {
			t.Error("Fetch() made a request despite the host not matching any rule")
		}
	})
}

func TestFetcherCustomExtract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>not an article readability could handle</html>"))
	}))
	defer srv.Close()

	var gotURL, gotSource string
	custom := func(html []byte, articleURL, source string) (*Article, error) {
		gotURL, gotSource = articleURL, source
		return &Article{URL: articleURL, Title: "Custom", Content: string(html), Source: source}, nil
	}
	rules := []HostRule{{Source: "Test", Hosts: []string{"127.0.0.1"}, Extract: custom}}
	f := NewFetcher(srv.Client(), rules)

	url := srv.URL + "/story"
	art, err := f.Fetch(context.Background(), url)
	if err != nil {
		t.Fatalf("Fetch() error = %v, want the custom extractor's article", err)
	}
	if art.Title != "Custom" {
		t.Errorf("Title = %q, want the custom extractor's article", art.Title)
	}
	if gotURL != url || gotSource != "Test" {
		t.Errorf("custom extractor called with (%q, %q), want (%q, %q)", gotURL, gotSource, url, "Test")
	}
}
