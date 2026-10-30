package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"golang.org/x/net/html/charset"
)

// Parse decodes an RSS 2.0 or Atom document. Atom feeds are converted to RSS
// form. Documents in encodings other than UTF-8 are converted using their XML
// declaration.
func Parse(data []byte) (RSS, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = charset.NewReaderLabel

	root, err := rootElement(dec)
	if err != nil {
		return RSS{}, err
	}

	switch root.Name.Local {
	case "rss":
		var feed RSS
		if err := dec.DecodeElement(&feed, &root); err != nil {
			return RSS{}, err
		}
		return feed, nil
	case "feed":
		var feed atomFeed
		if err := dec.DecodeElement(&feed, &root); err != nil {
			return RSS{}, err
		}
		return feed.toRSS(), nil
	default:
		return RSS{}, fmt.Errorf("unsupported feed format <%s>", root.Name.Local)
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

// atomFeed is the subset of an Atom (RFC 4287) document knews uses.
type atomFeed struct {
	Title   string      `xml:"title"`
	Link    []atomLink  `xml:"link"`
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

func (f atomFeed) toRSS() RSS {
	items := make([]Item, len(f.Entries))
	for i, e := range f.Entries {
		pubDate := e.Published
		if pubDate == "" {
			pubDate = e.Updated
		}
		authors := make([]string, len(e.Authors))
		for j, a := range e.Authors {
			authors[j] = a.Name
		}
		items[i] = Item{
			Title:       e.Title,
			Description: e.Summary,
			PubDate:     pubDate,
			Link:        alternateLink(e.Links),
			GUID:        e.ID,
			Author:      strings.Join(authors, ", "),
		}
	}
	return RSS{Channel: Channel{Title: f.Title, Link: alternateLink(f.Link), Items: items}}
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
