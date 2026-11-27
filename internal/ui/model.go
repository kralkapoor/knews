// Package ui implements the knews terminal interface with Bubble Tea.
package ui

import (
	"context"
	"strings"

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

// Model is the root bubbletea model for the application. It holds the state
// of every screen; only activeScreen is shown at a time.
type Model struct {
	activeScreen screen
	helpReturn   screen // screen to go back to when help closes
	width        int    // terminal size
	height       int

	// Splash screen: the providers on offer and the highlighted one.
	providers    []config.Provider
	providerList listCursor

	// The provider the user picked; zero until they pick one.
	provider config.Provider

	main    mainState
	article articleState

	feeds    feedSource
	articles articleSource

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
		providers:    providers,
		main:         newMainState(),
		article:      newArticleState(),
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
		m.width = msg.Width
		m.height = msg.Height
		m.syncMainScroll()
		m.resizeArticleViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	// Results of background work, handled whichever screen is showing.
	case feedsMsg:
		if msg.reqID != m.feedReqID {
			return m, nil
		}
		m.main.loading = false
		m.setFeedResults(msg.results)
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
		if msg.err != nil || msg.article == nil {
			// Fall back to what the feed says about the article.
			m.showArticleSummary(summaryArticle(msg.item, m.provider.Name), msg.err)
			return m, nil
		}
		m.showArticle(msg.article, "")
		return m, nil

	case openURLMsg:
		if m.activeScreen == screenArticle {
			return m, m.updateArticle(msg)
		}
	}
	return m, nil
}

// handleKey routes a key press: first the keys that work on every screen,
// then the keys of the active screen.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Any key closes the help screen.
	if m.activeScreen == screenHelp {
		m.activeScreen = m.helpReturn
		return m, nil
	}

	if msg.String() == "?" {
		// Like any other key, "?" clears the screen's transient notice.
		m.main.notice = ""
		m.article.notice = ""
		m.helpReturn = m.activeScreen
		m.activeScreen = screenHelp
		return m, nil
	}

	var cmd tea.Cmd
	switch m.activeScreen {
	case screenSplash:
		cmd = m.updateSplash(msg.String())
	case screenMain:
		cmd = m.updateMain(msg.String())
	case screenArticle:
		cmd = m.updateArticle(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.activeScreen {
	case screenSplash:
		return m.viewSplash()
	case screenMain:
		return m.viewMain()
	case screenArticle:
		return m.viewArticle()
	case screenHelp:
		return renderHelpView(m.width, m.height)
	}
	return ""
}

// selectProvider switches to the main screen for p and starts fetching its
// feeds.
func (m *Model) selectProvider(p config.Provider) tea.Cmd {
	m.provider = p
	m.main = newMainState()
	m.activeScreen = screenMain
	return m.fetchFeeds(false)
}

// openArticle switches to the article screen for item. Full-text providers
// start fetching the page; others show the feed's summary straight away.
func (m *Model) openArticle(item rss.Item) tea.Cmd {
	m.activeScreen = screenArticle
	if !m.provider.FullText {
		m.showArticleSummary(summaryArticle(item, m.provider.Name), nil)
		return nil
	}
	m.showArticleLoading(item)
	return m.fetchArticle(item)
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
	return fetchArticleCmd(ctx, m.articles, item, m.articleReqID)
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

// summaryArticle builds an article from what the feed says about it, for
// providers whose pages can't be extracted or when extraction fails.
func summaryArticle(item rss.Item, source string) *article.Article {
	published, _ := item.Published()
	return &article.Article{
		URL:         item.Link,
		Title:       strings.TrimSpace(item.Title),
		Author:      strings.TrimSpace(item.Author),
		Content:     item.Summary(),
		PublishedAt: published,
		Source:      source,
	}
}
