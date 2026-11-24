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

	feedFetcher := rss.NewFetcher(httpClient)

	// Only providers whose article pages can be extracted get full text.
	var rules []article.HostRule
	for _, p := range providers {
		if p.FullText {
			rules = append(rules, article.HostRule{Source: p.Name, Hosts: p.Hosts})
		}
	}
	articleFetcher := article.NewFetcher(httpClient, rules)

	model := ui.NewModel(providers, feedFetcher, articleFetcher)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
