package rss

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newFeedServer serves the RNZ fixture at /feed, optionally after ?delay=<ms>,
// invalid XML at /bad, and 500s at /fail. It counts requests in hits.
func newFeedServer(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	fixture, err := os.ReadFile("testdata/rnz-politics.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if ms, err := strconv.Atoi(r.URL.Query().Get("delay")); err == nil {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
		switch r.URL.Path {
		case "/feed":
			w.Write(fixture)
		case "/bad":
			w.Write([]byte("<html>not a feed</html>"))
		default:
			http.Error(w, "boom", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestClientFetch(t *testing.T) {
	var hits atomic.Int32
	srv := newFeedServer(t, &hits)
	fetcher := NewFetcher(srv.Client())

	tests := []struct {
		path    string
		wantErr string
	}{
		{path: "/feed"},
		{path: "/fail", wantErr: "500"},
		{path: "/bad", wantErr: "parse feed"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			feed, err := fetcher.Fetch(context.Background(), srv.URL+tt.path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Fetch() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}
			if len(feed.Items) != 16 {
				t.Errorf("got %d items, want 16", len(feed.Items))
			}
		})
	}
}

func TestFetchAllEmpty(t *testing.T) {
	f := NewFetcher(nil)
	if got := f.FetchAll(context.Background(), nil); got != nil {
		t.Errorf("FetchAll(nil) = %v, want nil", got)
	}
}

func TestFetchAllPreservesOrder(t *testing.T) {
	var hits atomic.Int32
	srv := newFeedServer(t, &hits)
	f := NewFetcher(srv.Client())
	f.MaxWorkers = 3

	// Earlier requests finish last.
	requests := []FeedRequest{
		{Name: "a", URL: srv.URL + "/feed?delay=60"},
		{Name: "b", URL: srv.URL + "/fail?delay=30"},
		{Name: "c", URL: srv.URL + "/feed"},
	}
	results := f.FetchAll(context.Background(), requests)

	if len(results) != len(requests) {
		t.Fatalf("got %d results, want %d", len(results), len(requests))
	}
	for i, r := range results {
		if r.FeedRequest != requests[i] {
			t.Errorf("results[%d] is for %+v, want %+v", i, r.FeedRequest, requests[i])
		}
		wantErr := r.Name == "b"
		if (r.Err != nil) != wantErr {
			t.Errorf("results[%d].Err = %v, want error: %v", i, r.Err, wantErr)
		}
	}
}

func TestFetchAllLimitsConcurrency(t *testing.T) {
	const maxWorkers = 3
	var inFlight, peak atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		w.Write([]byte("<rss><channel/></rss>"))
	}))
	defer srv.Close()

	requests := make([]FeedRequest, 10)
	for i := range requests {
		requests[i] = FeedRequest{Name: strconv.Itoa(i), URL: srv.URL}
	}
	f := NewFetcher(srv.Client())
	f.MaxWorkers = maxWorkers
	f.FetchAll(context.Background(), requests)

	if got := peak.Load(); got > maxWorkers {
		t.Errorf("peak concurrency = %d, want <= %d", got, maxWorkers)
	}
}

func TestFetchAllCancelled(t *testing.T) {
	var hits atomic.Int32
	srv := newFeedServer(t, &hits)
	f := NewFetcher(srv.Client())
	f.MaxWorkers = 2

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	requests := []FeedRequest{
		{Name: "a", URL: srv.URL + "/feed"},
		{Name: "b", URL: srv.URL + "/feed"},
		{Name: "c", URL: srv.URL + "/feed"},
		{Name: "d", URL: srv.URL + "/feed"},
	}
	results := f.FetchAll(ctx, requests)

	for i, r := range results {
		if r.FeedRequest != requests[i] {
			t.Errorf("results[%d] is for %+v, want %+v", i, r.FeedRequest, requests[i])
		}
		if !errors.Is(r.Err, context.Canceled) {
			t.Errorf("results[%d].Err = %v, want context.Canceled", i, r.Err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server got %d requests, want 0", n)
	}
}
