package main

import (
	"log"

	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/kralkapoor/knews/internal/ui"
)

const maxWorkers = 5

func main() {
	// Initialize RSS feed client and fetcher.
	// FeedClient handles individual HTTP requests, FeedFetcher coordinates
	// fetching multiple feeds concurrently using a worker pool.
	feedClient := rss.NewFeedClient()
	feedFetcher := rss.NewFeedFetcher(feedClient, maxWorkers)

	// Initialize article parser registry with RNZ parser
	parserRegistry := article.NewParserRegistry()
	parserRegistry.Register(article.NewRNZParser())

	// Initialize article fetcher
	articleFetcher := article.NewArticleFetcher(parserRegistry)

	// Create and run the UI application
	app := ui.NewApp(feedFetcher, articleFetcher)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
