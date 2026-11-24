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

	if got, want := feed.Title, "RNZ Politics Headlines"; got != want {
		t.Errorf("channel title = %q, want %q", got, want)
	}
	if got, want := len(feed.Items), 16; got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}

	first := feed.Items[0]
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

func TestParseAtom(t *testing.T) {
	data, err := os.ReadFile("testdata/stuff-nz-news.xml")
	if err != nil {
		t.Fatal(err)
	}

	feed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got, want := feed.Title, "Stuff - Nz News"; got != want {
		t.Errorf("channel title = %q, want %q", got, want)
	}
	if got, want := len(feed.Items), 3; got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}

	const link = "https://www.stuff.co.nz/business/361053141/tattoo-parlor-boss-told-traumatised-worker-harden-and-threatened-call-immigration"
	want := Item{
		Title:       "Boss ordered to pay $47k after visa threats and abusive texts",
		Description: "A Wellington employer who threatened to report a migrant tattoo artist to immigration after he took sick leave must pay $40,000 in compensation.",
		PubDate:     "2026-10-30T00:47:20Z",
		Link:        link,
		GUID:        link,
		Author:      "Stuff",
	}
	if got := feed.Items[0]; got != want {
		t.Errorf("first item =\n%#v\nwant\n%#v", got, want)
	}
	if _, ok := feed.Items[0].Published(); !ok {
		t.Error("Atom published date didn't parse")
	}
}

func TestAtomEntryFallbacks(t *testing.T) {
	data := []byte(`<feed xmlns="http://www.w3.org/2005/Atom"><entry>
		<link rel="self" href="https://example.com/self"/>
		<link href="https://example.com/story"/>
		<updated>2026-10-30T00:00:00Z</updated>
		<author><name>A</name></author><author><name>B</name></author>
	</entry></feed>`)

	feed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	item := feed.Items[0]
	if item.Link != "https://example.com/story" {
		t.Errorf("Link = %q, want the alternate link", item.Link)
	}
	if item.PubDate != "2026-10-30T00:00:00Z" {
		t.Errorf("PubDate = %q, want the updated date when published is missing", item.PubDate)
	}
	if item.Author != "A, B" {
		t.Errorf("Author = %q, want %q", item.Author, "A, B")
	}
}

func TestParseRSSAuthor(t *testing.T) {
	data, err := os.ReadFile("testdata/herald-nz.xml")
	if err != nil {
		t.Fatal(err)
	}

	feed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	items := feed.Items
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	for i, want := range []string{"Joel Kulasingham", "", "Doug Laing"} {
		if items[i].Author != want {
			t.Errorf("items[%d].Author = %q, want %q", i, items[i].Author, want)
		}
	}
	if got, want := items[0].Title, "Greyhound charity apologises after claiming dog scared of men needed lesbian home"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestItemSummary(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Plain text.", "Plain text."},
		{"\n        Padded and\n  wrapped.\n      ", "Padded and wrapped."},
		{"<p>First.</p><p>Second &amp; <b>bold</b>.</p>", "First. Second & bold."},
		{"Line<br/>break", "Line break"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := (Item{Description: tt.in}).Summary(); got != tt.want {
			t.Errorf("Summary(%q) = %q, want %q", tt.in, got, tt.want)
		}
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
		{name: "unsupported root", input: `<html><body>not a feed</body></html>`},
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
	if got := feed.Items[0].Title; got != "Café" {
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
