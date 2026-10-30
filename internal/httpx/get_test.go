package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.UserAgent()
		switch r.URL.Path {
		case "/ok":
			w.Write([]byte("hello"))
		case "/huge":
			w.Write(make([]byte, MaxBodyBytes+1))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr string
	}{
		{name: "ok", path: "/ok", want: "hello"},
		{name: "not found", path: "/missing", wantErr: "404"},
		{name: "body too large", path: "/huge", wantErr: "exceeds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := srv.URL + tt.path
			body, err := Get(context.Background(), srv.Client(), url)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || !strings.Contains(err.Error(), url) {
					t.Fatalf("Get() error = %v, want error mentioning %q and the URL", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if string(body) != tt.want {
				t.Errorf("Get() = %q, want %q", body, tt.want)
			}
			if gotUA != UserAgent {
				t.Errorf("User-Agent = %q, want %q", gotUA, UserAgent)
			}
		})
	}
}

func TestGetHonoursContextDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := Get(ctx, srv.Client(), srv.URL)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get() error = %v, want context.DeadlineExceeded", err)
	}
}
