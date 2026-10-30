// Package ui implements the knews terminal interface with Bubble Tea.
package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
)

type screen int

const (
	screenSplash screen = iota
	screenMain
	screenArticle
	screenHelp
)

// Model is the root bubbletea model for the application.
type Model struct {
	activeScreen screen
	helpReturn   screen // screen to go back to when help closes
	splash       splashModel
	main         mainModel
	articleView  articleModel
	help         helpModel

	feeds    feedSource
	articles articleSource
	provider config.Provider // zero until the user picks one

	// Request IDs of the latest feed and article fetches. Results carrying
	// any other ID are stale and dropped.
	feedReqID     int
	articleReqID  int
	cancelArticle context.CancelFunc // cancels the in-flight article fetch, if any
}

// NewModel returns the root model. providers are offered on the splash
// screen; feeds and articles are usually an *rss.Fetcher and an
// *article.Fetcher.
func NewModel(providers []config.Provider, feeds feedSource, articles articleSource) Model {
	return Model{
		activeScreen: screenSplash,
		splash:       newSplashModel(providers),
		main:         newMainModel(),
		articleView:  newArticleModel(),
		feeds:        feeds,
		articles:     articles,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.splash.setSize(msg.Width, msg.Height)
		m.main.setSize(msg.Width, msg.Height)
		m.articleView.setSize(msg.Width, msg.Height)
		m.help.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	// Async results — handled regardless of the active screen.
	case feedsMsg:
		if msg.reqID != m.feedReqID {
			return m, nil
		}
		m.main.loading = false
		m.main.setFeedResults(msg.results)
		if msg.refresh {
			m.main.notice = successStyle.Render("Feeds refreshed")
		}
		return m, nil

	case articleMsg:
		if msg.reqID != m.articleReqID {
			return m, nil
		}
		if m.cancelArticle != nil {
			m.cancelArticle()
			m.cancelArticle = nil
		}
		switch {
		case msg.err != nil:
			m.articleView.showError(msg.err)
		case msg.article != nil:
			m.articleView.showArticle(msg.article)
		}
		return m, nil

	// Screen transitions.
	case providerSelectedMsg:
		m.provider = msg.provider
		m.main.setFeeds(msg.provider.Feeds)
		m.activeScreen = screenMain
		return m, m.fetchFeeds(false)

	case requestRefreshMsg:
		if m.main.loading {
			return m, nil
		}
		return m, m.fetchFeeds(true)

	case openArticleMsg:
		m.articleView.showLoading(msg.item)
		m.activeScreen = screenArticle
		return m, m.fetchArticle(msg.item)

	case closeArticleMsg:
		m.stopArticleFetch()
		m.activeScreen = screenMain
		return m, nil

	case showHelpMsg:
		m.helpReturn = m.activeScreen
		m.activeScreen = screenHelp
		return m, nil

	case closeHelpMsg:
		m.activeScreen = m.helpReturn
		return m, nil
	}

	// Route everything else to the active screen.
	var cmd tea.Cmd
	switch m.activeScreen {
	case screenSplash:
		m.splash, cmd = m.splash.Update(msg)
	case screenMain:
		m.main, cmd = m.main.Update(msg)
	case screenArticle:
		m.articleView, cmd = m.articleView.Update(msg)
	case screenHelp:
		m.help, cmd = m.help.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.activeScreen {
	case screenSplash:
		return m.splash.View()
	case screenMain:
		return m.main.View()
	case screenArticle:
		return m.articleView.View()
	case screenHelp:
		return m.help.View()
	}
	return ""
}

// fetchFeeds starts fetching the selected provider's feeds, superseding any
// fetch already in flight.
func (m *Model) fetchFeeds(refresh bool) tea.Cmd {
	m.feedReqID++
	m.main.loading = true
	return fetchFeedsCmd(m.feeds, m.provider.Feeds, m.feedReqID, refresh)
}

// fetchArticle starts fetching item's article, cancelling any fetch already
// in flight.
func (m *Model) fetchArticle(item rss.Item) tea.Cmd {
	m.stopArticleFetch()
	ctx, cancel := context.WithTimeout(context.Background(), articleTimeout)
	m.cancelArticle = cancel
	return fetchArticleCmd(ctx, m.articles, item.Link, m.articleReqID)
}

// stopArticleFetch cancels the in-flight article fetch, if any, and makes
// sure its result is ignored.
func (m *Model) stopArticleFetch() {
	if m.cancelArticle != nil {
		m.cancelArticle()
		m.cancelArticle = nil
	}
	m.articleReqID++
}
