package rss

import (
	"fmt"
	"io"
	"net/http"
)

func GetUriBytes(uri string) []byte {
	resp, err := http.Get(uri)
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
	// fmt.Println(string(bodyBytes))
	return bodyBytes
}

// Parse bytes from http request into http article struct



