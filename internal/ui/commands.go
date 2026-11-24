package ui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
)

const (
	feedsTimeout   = 30 * time.Second
	articleTimeout = 15 * time.Second
)

// feedSource fetches RSS feeds; *rss.Fetcher implements it. It is an
// interface only so the tests can swap in a fake.
type feedSource interface {
	FetchAll(ctx context.Context, requests []rss.FeedRequest) []rss.FeedResult
}

// articleSource fetches full articles; *article.Fetcher implements it. It is
// an interface only so the tests can swap in a fake.
type articleSource interface {
	Fetch(ctx context.Context, url string) (*article.Article, error)
}

// Messages returned by the commands below, which do slow work (network
// requests, launching a browser) in the background. The fetch results carry
// the request ID they were started with so that results the user has since
// moved on from can be dropped.

type feedsMsg struct {
	reqID   int
	refresh bool
	results []rss.FeedResult
}

type articleMsg struct {
	reqID   int
	item    rss.Item
	article *article.Article
	err     error
}

type openURLMsg struct {
	err error
}

// fetchFeedsCmd fetches all of a provider's feeds.
func fetchFeedsCmd(src feedSource, feeds []config.Feed, reqID int, refresh bool) tea.Cmd {
	requests := make([]rss.FeedRequest, len(feeds))
	for i, f := range feeds {
		requests[i] = rss.FeedRequest{Name: f.Name, URL: f.URL}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), feedsTimeout)
		defer cancel()
		return feedsMsg{reqID: reqID, refresh: refresh, results: src.FetchAll(ctx, requests)}
	}
}

// fetchArticleCmd fetches the article item links to. The caller owns ctx so
// it can cancel the fetch when the user leaves the article view.
func fetchArticleCmd(ctx context.Context, src articleSource, item rss.Item, reqID int) tea.Cmd {
	return func() tea.Msg {
		art, err := src.Fetch(ctx, item.Link)
		return articleMsg{reqID: reqID, item: item, article: art, err: err}
	}
}

// openURLCmd opens url in the default browser.
func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			// Unlike "cmd /c start", this doesn't treat & in the URL as a
			// command separator.
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		return openURLMsg{err: cmd.Run()}
	}
}
