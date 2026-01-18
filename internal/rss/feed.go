package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// get collection of rss feeds, e.g. business, sports etc

func listFeeds() []string {
	feeds := []string{"https://www.rnz.co.nz/rss/political"}
	fmt.Printf("Configured feeds are: %v\n***\n\n",feeds)
	return feeds
}

// http get request to the target feed
func getFeed(feed string) []byte {
	resp, err := http.Get(feed)
	if err != nil {
		fmt.Println("could not get the feed for some reason")
		fmt.Println(err)
	}
	
	// defer closing the resp body to ensure closure and prevent leaks
	defer resp.Body.Close()

	// check the request status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("response err not ok: %v", resp.StatusCode)
	}
	
	// read response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)

	// todo handle any errors if it cannot reach eof

	// convert the byte slice into a string so that it's human readable
	bodyString := bodyBytes
	fmt.Println(string(bodyString))
	return bodyString
}

// get xml from response

func parseXmlResponse(resp []byte) (Rss, error) {

	// assign new rss var of type Rss
	// it's defined separately because we will unmarshal the xml data slice into it via memory pointer
	var rss Rss
	err := xml.Unmarshal(resp, &rss)
	
	if err != nil {
		return Rss{}, err
	}

	return rss, nil
}

// parse xml into a struct

// print out information from the struct
