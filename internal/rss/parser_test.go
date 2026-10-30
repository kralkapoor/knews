package rss

import (
	"encoding/xml"
	"errors"
	"os"
	"testing"
	"time"
)

func TestParseFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/rnz-politics.xml")
	if err != nil {
		t.Fatal(err)
	}

	feed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := feed.Channel.Title, "RNZ Politics Headlines"; got != want {
		t.Errorf("channel title = %q, want %q", got, want)
	}
	if got, want := len(feed.Channel.Items), 16; got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}

	first := feed.Channel.Items[0]
	want := Item{
		Title:       "Government seeks urgent advice on how to protect Auckland rockpools",
		Description: "\n        Whangaparāoa residents claim visitors have stripped local rockpools bare of sealife.\n      ",
		PubDate:     "Sun, 18 Jan 2026 19:25:50 +1300",
		Link:        "https://www.rnz.co.nz/news/political/584370/government-seeks-urgent-advice-on-how-to-protect-auckland-rockpools",
		GUID:        "05a72d4bc4026f8039e7417f6a56dfc86c099a36",
	}
	if first != want {
		t.Errorf("first item =\n%#v\nwant\n%#v", first, want)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		syntax bool // expect an *xml.SyntaxError
	}{
		{name: "empty", input: ""},
		{name: "malformed", input: "<rss><channel><item>", syntax: true},
		{name: "atom root", input: `<feed xmlns="http://www.w3.org/2005/Atom"><entry/></feed>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.input))
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
			var syntaxErr *xml.SyntaxError
			if tt.syntax && !errors.As(err, &syntaxErr) {
				t.Errorf("Parse() error = %v, want *xml.SyntaxError", err)
			}
		})
	}
}

func TestParseNonUTF8(t *testing.T) {
	data := []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>" +
		"<rss><channel><item><title>Caf\xe9</title></item></channel></rss>")

	feed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := feed.Channel.Items[0].Title; got != "Café" {
		t.Errorf("title = %q, want %q", got, "Café")
	}
}

func TestItemPublished(t *testing.T) {
	nzdt := time.FixedZone("", 13*60*60)
	tests := []struct {
		pubDate string
		want    time.Time
		ok      bool
	}{
		{"Sun, 18 Jan 2026 19:25:50 +1300", time.Date(2026, 1, 18, 19, 25, 50, 0, nzdt), true},
		{"Fri, 2 Jan 2026 08:00:00 +1300", time.Date(2026, 1, 2, 8, 0, 0, 0, nzdt), true},
		{"Fri, 02 Jan 2026 08:00:00 GMT", time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC), true},
		{"2026-10-30T00:21:36Z", time.Date(2026, 10, 30, 0, 21, 36, 0, time.UTC), true},
		{"2026-10-30", time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC), true},
		{"", time.Time{}, false},
		{"last Tuesday", time.Time{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.pubDate, func(t *testing.T) {
			got, ok := Item{PubDate: tt.pubDate}.Published()
			if ok != tt.ok || !got.Equal(tt.want) {
				t.Errorf("Published() = %v, %v; want %v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}
