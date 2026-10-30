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

// stubParser handles URLs containing prefix and returns an article titled name.
type stubParser struct {
	prefix string
	name   string
}

func (s stubParser) CanParse(url string) bool { return strings.HasPrefix(url, s.prefix) }

func (s stubParser) Parse(_ []byte, url string) (*Article, error) {
	return &Article{URL: url, Title: s.name}, nil
}

func TestRegistryParserFor(t *testing.T) {
	var r Registry
	r.Register(stubParser{prefix: "https://a.example/news", name: "specific"})
	r.Register(stubParser{prefix: "https://a.example", name: "general"})

	tests := []struct {
		url  string
		want string // parser name, or "" for none
	}{
		{"https://a.example/news/1", "specific"},
		{"https://a.example/sport/1", "general"},
		{"https://b.example/news/1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			p := r.parserFor(tt.url)
			got := ""
			if p != nil {
				got = p.(stubParser).name
			}
			if got != tt.want {
				t.Errorf("parserFor(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
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

	var registry Registry
	registry.Register(ReadabilityParser{Source: "Test", Hosts: []string{"127.0.0.1"}})
	f := NewFetcher(&registry, srv.Client())

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

	t.Run("no parser", func(t *testing.T) {
		before := hits.Load()
		_, err := f.Fetch(context.Background(), "https://unknown.example/story")
		if !errors.Is(err, ErrNoParser) {
			t.Fatalf("Fetch() error = %v, want ErrNoParser", err)
		}
		if hits.Load() != before {
			t.Error("Fetch() made a request despite having no parser")
		}
	})
}
