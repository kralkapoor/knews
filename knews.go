package main

import (
	"log"
	"time"

	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/kralkapoor/knews/internal/ui"
)

const (
	feedTimeout    = 10 * time.Second
	articleTimeout = 15 * time.Second
	maxWorkers     = 5
)

func main() {
	// Initialize RSS feed client and fetcher
	feedClient := rss.NewFeedClient(feedTimeout)
	feedFetcher := rss.NewFeedFetcher(feedClient, maxWorkers)

	// Initialize article parser registry with RNZ parser
	parserRegistry := article.NewParserRegistry()
	parserRegistry.Register(article.NewRNZParser())

	// Initialize article fetcher
	articleFetcher := article.NewArticleFetcher(parserRegistry, articleTimeout)

	// Create and run the UI application
	app := ui.NewApp(feedFetcher, articleFetcher)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
