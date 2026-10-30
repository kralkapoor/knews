// Command knews is a terminal reader for New Zealand news feeds.
package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/httpx"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/kralkapoor/knews/internal/ui"
)

func main() {
	providers := config.Providers()
	httpClient := httpx.NewClient()

	feeds := rss.NewFetcher(rss.NewClient(httpClient), 0) // 0: default worker count

	var registry article.Registry
	for _, p := range providers {
		if p.FullText {
			registry.Register(article.ReadabilityParser{Source: p.Name, Hosts: p.Hosts})
		}
	}
	articles := article.NewFetcher(&registry, httpClient)

	p := tea.NewProgram(ui.NewModel(providers, feeds, articles), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
