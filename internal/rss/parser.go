package rss

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

// Parse decodes an RSS 2.0 document. Documents in encodings other than UTF-8
// are converted using their XML declaration.
func Parse(data []byte) (RSS, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.CharsetReader = charset.NewReaderLabel

	var feed RSS
	if err := dec.Decode(&feed); err != nil {
		return RSS{}, err
	}
	return feed, nil
}
