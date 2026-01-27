package rss

import "encoding/xml"

// parse xml into a struct
func ParseXmlResponse(resp []byte) (Rss, error) {
	var rss Rss

	err := xml.Unmarshal(resp, &rss)

	if err != nil {
		return Rss{}, err
	}

	return rss, err
}
