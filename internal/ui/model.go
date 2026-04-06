package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kralkapoor/knews/internal/article"
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
	splash       splashModel
	main         mainModel
	articleView  articleModel
	help         helpModel

	feedFetcher      *rss.FeedFetcher
	articleFetcher   *article.ArticleFetcher
	selectedProvider string
	width            int
	height           int
}

// NewModel creates a new root Model with all sub-models initialized.
func NewModel(feedFetcher *rss.FeedFetcher, articleFetcher *article.ArticleFetcher) Model {
	return Model{
		activeScreen:   screenSplash,
		splash:         newSplashModel(),
		main:           newMainModel(),
		articleView:    newArticleModel(),
		help:           newHelpModel(),
		feedFetcher:    feedFetcher,
		articleFetcher: articleFetcher,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.splash.width = msg.Width
		m.splash.height = msg.Height
		m.splash.providerList.SetSize(30, msg.Height/3)
		m.main.setSize(msg.Width, msg.Height)
		m.articleView.setSize(msg.Width, msg.Height)
		m.help.width = msg.Width
		m.help.height = msg.Height
		return m, nil

	// Async result messages — handled regardless of active screen.
	case feedsLoadedMsg:
		feedOrder := config.GetProvider(m.selectedProvider).FeedOrder
		firstFeed := ""
		if len(feedOrder) > 0 {
			firstFeed = feedOrder[0]
		}
		m.main.setFeedResults(msg.results, firstFeed)
		m.main.statusMsg = statusText("Press ", "?", " for help │ ", "q", " to quit")
		return m, nil

	case feedsRefreshedMsg:
		m.main.setFeedResults(msg.results, m.main.currentFeed)
		m.main.statusMsg = successStyle.Render("Feeds refreshed") +
			statusText(" │ Press ", "?", " for help")
		return m, nil

	case articleLoadedMsg:
		m.articleView.showArticle(msg.article)
		return m, nil

	case articleErrorMsg:
		m.articleView.showError(msg.title, msg.err)
		return m, nil

	// Screen transition messages.
	case providerSelectedMsg:
		m.selectedProvider = msg.name
		provider := config.GetProvider(msg.name)
		m.main.setFeeds(provider.FeedOrder)
		m.main.statusMsg = keyStyle.Render("Loading feeds...")
		m.activeScreen = screenMain
		return m, fetchFeedsCmd(m.feedFetcher, provider)

	case openArticleMsg:
		m.articleView.showLoading(msg.item.Title)
		m.activeScreen = screenArticle
		return m, fetchArticleCmd(m.articleFetcher, msg.item)

	case showHelpMsg:
		m.activeScreen = screenHelp
		return m, nil

	case closeArticleMsg:
		m.activeScreen = screenMain
		return m, nil

	case closeHelpMsg:
		m.activeScreen = screenMain
		return m, nil

	case requestRefreshMsg:
		if m.selectedProvider == "" {
			return m, nil
		}
		provider := config.GetProvider(m.selectedProvider)
		m.main.statusMsg = keyStyle.Render("Refreshing feeds...")
		return m, refreshFeedsCmd(m.feedFetcher, provider)
	}

	// Route to active screen's Update.
	switch m.activeScreen {
	case screenSplash:
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		return m, cmd

	case screenMain:
		var cmd tea.Cmd
		m.main, cmd = m.main.Update(msg)
		return m, cmd

	case screenArticle:
		var cmd tea.Cmd
		m.articleView, cmd = m.articleView.Update(msg)
		return m, cmd

	case screenHelp:
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd
	}

	return m, nil
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
