package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
)

// Messages returned by async commands.

type feedsLoadedMsg struct {
	results []rss.FeedResult
}

type feedsRefreshedMsg struct {
	results []rss.FeedResult
}

type articleLoadedMsg struct {
	article *article.Article
	title   string
}

type articleErrorMsg struct {
	err   error
	title string
}

// fetchFeedsCmd fetches all RSS feeds for a provider (30s timeout).
func fetchFeedsCmd(fetcher *rss.FeedFetcher, provider config.Provider) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results := fetcher.FetchAllFromMap(ctx, provider.Feeds)
		return feedsLoadedMsg{results: results}
	}
}

// refreshFeedsCmd re-fetches all RSS feeds for a provider (30s timeout).
func refreshFeedsCmd(fetcher *rss.FeedFetcher, provider config.Provider) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results := fetcher.FetchAllFromMap(ctx, provider.Feeds)
		return feedsRefreshedMsg{results: results}
	}
}

// fetchArticleCmd fetches full article content (15s timeout).
func fetchArticleCmd(fetcher *article.ArticleFetcher, item rss.Item) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		art, err := fetcher.Fetch(ctx, item.Link)
		if err != nil {
			return articleErrorMsg{err: err, title: item.Title}
		}
		return articleLoadedMsg{article: art, title: item.Title}
	}
}

// Internal messages for screen transitions.

type providerSelectedMsg struct {
	name string
}

type openArticleMsg struct {
	item rss.Item
}

type showHelpMsg struct{}

type closeArticleMsg struct{}

type closeHelpMsg struct{}

type requestRefreshMsg struct{}
