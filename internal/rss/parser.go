package rss

import "encoding/xml"

// parse xml into a struct
func ParseXmlResponse(resp []byte) (Rss, error) {

	// assign new rss var of type Rss
	// it's defined separately because we will unmarshal the xml data slice into it via memory pointer
	var rss Rss
	err := xml.Unmarshal(resp, &rss)
	
	if err != nil {
		return Rss{}, err
	}

	return rss, err
}