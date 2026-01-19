package main

import (
	"fmt"
	"log"

	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
)

var feedConfiguration = config.GetTestFeedConfiguration()

func main() {
	fmt.Println("What's going on out there?")

	// get xml response as bytes
	testFeedUri := feedConfiguration.Providers[0].Feeds[0]
	xmlBytes := rss.GetUriBytes(testFeedUri)
	
	// unmarshal xml response into a struct
	rss, error := rss.ParseXmlResponse(xmlBytes)
	if error != nil {
		log.Fatal(error)
	}

	fmt.Println(rss)
	
}
