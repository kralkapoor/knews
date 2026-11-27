package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"golang.org/x/net/html/charset"
)

// Parse decodes an RSS 2.0 or Atom document. Documents in encodings other
// than UTF-8 are converted using their XML declaration.
func Parse(data []byte) (Feed, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = charset.NewReaderLabel

	root, err := rootElement(dec)
	if err != nil {
		return Feed{}, err
	}

	switch root.Name.Local {
	case "rss":
		var doc rssDocument
		if err := dec.DecodeElement(&doc, &root); err != nil {
			return Feed{}, err
		}
		return doc.toFeed(), nil
	case "feed":
		var doc atomFeed
		if err := dec.DecodeElement(&doc, &root); err != nil {
			return Feed{}, err
		}
		return doc.toFeed(), nil
	default:
		// Any other root is an error, so a feed in an unexpected format fails
		// loudly instead of showing as empty.
		return Feed{}, fmt.Errorf("unsupported feed format <%s>", root.Name.Local)
	}
}

// rootElement advances dec past the prolog to the document's root element.
func rootElement(dec *xml.Decoder) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start, nil
		}
	}
}

// ---- RSS 2.0 ----

// rssDocument is the root element of an RSS 2.0 document.
type rssDocument struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Author      string `xml:"http://purl.org/dc/elements/1.1/ creator"`
}

func (d rssDocument) toFeed() Feed {
	items := make([]Item, len(d.Channel.Items))
	for i, r := range d.Channel.Items {
		items[i] = Item{
			Title:       r.Title,
			Description: r.Description,
			PubDate:     r.PubDate,
			Link:        r.Link,
			GUID:        r.GUID,
			Author:      r.Author,
		}
	}
	return Feed{Title: d.Channel.Title, Items: items}
}

// ---- Atom (RFC 4287) ----

// atomFeed is the subset of an Atom document knews uses.
type atomFeed struct {
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string       `xml:"id"`
	Title     string       `xml:"title"`
	Links     []atomLink   `xml:"link"`
	Published string       `xml:"published"`
	Updated   string       `xml:"updated"`
	Summary   string       `xml:"summary"`
	Authors   []atomPerson `xml:"author"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomPerson struct {
	Name string `xml:"name"`
}

func (f atomFeed) toFeed() Feed {
	items := make([]Item, len(f.Entries))
	for i, e := range f.Entries {
		// Prefer the published date, but fall back to the updated date.
		pubDate := e.Published
		if pubDate == "" {
			pubDate = e.Updated
		}

		authorNames := make([]string, len(e.Authors))
		for j, a := range e.Authors {
			authorNames[j] = a.Name
		}

		items[i] = Item{
			Title:       e.Title,
			Description: e.Summary,
			PubDate:     pubDate,
			Link:        alternateLink(e.Links),
			GUID:        e.ID,
			Author:      strings.Join(authorNames, ", "),
		}
	}
	return Feed{Title: f.Title, Items: items}
}

// alternateLink returns the href of the link pointing at the entry's web
// page: the one with rel "alternate", which is the default when rel is absent.
func alternateLink(links []atomLink) string {
	for _, l := range links {
		if l.Rel == "" || l.Rel == "alternate" {
			return l.Href
		}
	}
	return ""
}
