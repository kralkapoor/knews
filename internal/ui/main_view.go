package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/rss"
)

// focusPanel represents which panel currently has focus.
type focusPanel int

const (
	focusFeedList focusPanel = iota
	focusArticleList
)

const (
	feedPanelWidth  = 22 // 20 inner + 2 for borders
	minArticleWidth = 10
	headerHeight    = 3
	statusHeight    = 1
)

// mainState is everything the two-panel feed/article screen remembers.
type mainState struct {
	focus       focusPanel
	feeds       listCursor                // the feed panel's cursor and scroll position
	articles    listCursor                // the article panel's cursor and scroll position
	feedResults map[string]rss.FeedResult // fetched feeds, by feed name
	loading     bool                      // a feed fetch is in flight
	notice      string                    // transient status message, cleared on the next key
}

func newMainState() mainState {
	return mainState{
		focus:       focusFeedList,
		feedResults: make(map[string]rss.FeedResult),
	}
}

// currentFeedName returns the name of the feed under the feed cursor, or ""
// if there are no feeds.
func (m Model) currentFeedName() string {
	if m.main.feeds.cursor >= len(m.provider.Feeds) {
		return ""
	}
	return m.provider.Feeds[m.main.feeds.cursor].Name
}

// currentArticles returns the items of the feed under the feed cursor. It is
// empty until that feed has been fetched.
func (m Model) currentArticles() []rss.Item {
	return m.main.feedResults[m.currentFeedName()].Feed.Items
}

// articlePanelWidth returns the width of the article panel, borders included.
func (m Model) articlePanelWidth() int {
	return max(m.width-feedPanelWidth, minArticleWidth+2)
}

// panelHeight returns the height of the two list panels, borders included.
func (m Model) panelHeight() int {
	return max(m.height-headerHeight-statusHeight, 3)
}

// feedRowsVisible returns how many feeds fit in the feed panel: one line
// each, inside the borders.
func (m Model) feedRowsVisible() int {
	return max(m.panelHeight()-2, 1)
}

// articleRowsVisible returns how many articles fit in the article panel: two
// lines each (title and date), inside the borders.
func (m Model) articleRowsVisible() int {
	return max((m.panelHeight()-2)/2, 1)
}

// syncMainScroll scrolls both panels just enough to keep their cursors
// visible. Call it whenever the lists change or the terminal is resized.
func (m *Model) syncMainScroll() {
	m.main.feeds.scroll(len(m.provider.Feeds), m.feedRowsVisible())
	m.main.articles.scroll(len(m.currentArticles()), m.articleRowsVisible())
}

// setFeedResults stores fetched feed data. The article cursor stays where it
// is unless the feed now has fewer articles than that.
func (m *Model) setFeedResults(results []rss.FeedResult) {
	for _, result := range results {
		m.main.feedResults[result.Name] = result
	}
	if m.main.articles.cursor >= len(m.currentArticles()) {
		m.main.articles = listCursor{}
	}
	m.syncMainScroll()
}

// mainStatus returns the text for the status bar.
func (m Model) mainStatus() string {
	if m.main.loading {
		return keyStyle.Render("Loading feeds...")
	}

	name := m.currentFeedName()
	if err := m.main.feedResults[name].Err; err != nil {
		return errorStyle.Render("Error loading "+name+": ") + err.Error()
	}

	hint := "Press " + keyHint("?", "for help") + " │ " + keyHint("Esc", "providers") + " │ " + keyHint("q", "to quit")
	if m.main.notice != "" {
		return m.main.notice + " │ " + hint
	}
	return hint
}

// updateMain handles a key press on the feed and article screen.
func (m *Model) updateMain(key string) tea.Cmd {
	m.main.notice = ""

	switch key {
	case "tab":
		if m.main.focus == focusFeedList {
			m.main.focus = focusArticleList
		} else {
			m.main.focus = focusFeedList
		}
	case "h", "left":
		m.main.focus = focusFeedList
	case "l", "right":
		m.main.focus = focusArticleList
	case "q":
		return tea.Quit
	case "esc", "backspace":
		// Go back to the provider list. Bumping the request ID drops the
		// results of any fetch still in flight.
		m.feedReqID++
		m.main.loading = false
		m.activeScreen = screenSplash
	case "r":
		if m.main.loading {
			return nil
		}
		return m.fetchFeeds(true)
	case "enter":
		if m.main.focus == focusFeedList {
			m.main.focus = focusArticleList
			return nil
		}
		articles := m.currentArticles()
		if m.main.articles.cursor < len(articles) {
			return m.openArticle(articles[m.main.articles.cursor])
		}
	default:
		// Anything else may be a navigation key for the focused list.
		if m.main.focus == focusArticleList {
			m.main.articles.handleKey(key, len(m.currentArticles()), m.articleRowsVisible())
			return nil
		}
		before := m.main.feeds.cursor
		m.main.feeds.handleKey(key, len(m.provider.Feeds), m.feedRowsVisible())
		if m.main.feeds.cursor != before {
			// A different feed: show its articles from the top.
			m.main.articles = listCursor{}
		}
	}
	return nil
}

// feedRow draws one line of the feed panel.
func feedRow(name string, selected bool, width int) string {
	style := lipgloss.NewStyle()
	if selected {
		style = selectedStyle.Bold(true)
	}
	return style.PaddingLeft(1).Width(width).Render(truncateString(name, width-1))
}

// articleRow draws two lines of the article panel: the title and the date.
func articleRow(item rss.Item, selected bool, width int) string {
	base := lipgloss.NewStyle()
	if selected {
		base = selectedStyle
	}
	titleLineStyle := base.Bold(selected).PaddingLeft(1)
	dateLineStyle := base.PaddingLeft(3)
	if !selected {
		dateLineStyle = dateLineStyle.Foreground(colorGray)
	}

	// Truncate rather than letting lipgloss wrap: a wrapped title would make
	// the row taller than two lines and push the list out of its box.
	title := titleLineStyle.Width(width).Render(truncateString(item.Title, width-1))
	date := dateLineStyle.Width(width).Render(truncateString(formatPubDate(item), width-3))
	return title + "\n" + date
}

func (m Model) viewMain() string {
	headerText := titleStyle.Render("Knews - " + m.provider.Name)
	headerContent := lipgloss.PlaceHorizontal(m.width-2, lipgloss.Center, headerText)
	header := titledBox(headerContent, "", m.width, headerHeight, colorDarkCyan)

	feedBorderColor := colorGray
	articleBorderColor := colorGray
	if m.main.focus == focusFeedList {
		feedBorderColor = colorDarkCyan
	} else {
		articleBorderColor = colorDarkCyan
	}

	articleTitle := "Articles"
	if m.currentFeedName() != "" {
		articleTitle = m.currentFeedName()
	}

	feedInnerWidth := feedPanelWidth - 2
	articleInnerWidth := m.articlePanelWidth() - 2

	// Feed rows: one line each, starting from the scroll offset.
	var feedRows []string
	feeds := m.provider.Feeds
	lastFeed := min(m.main.feeds.offset+m.feedRowsVisible(), len(feeds))
	for i := m.main.feeds.offset; i < lastFeed; i++ {
		feedRows = append(feedRows, feedRow(feeds[i].Name, i == m.main.feeds.cursor, feedInnerWidth))
	}
	if len(feedRows) == 0 {
		feedRows = []string{emptyListText}
	}

	// Article rows: two lines each, starting from the scroll offset.
	var articleRows []string
	articles := m.currentArticles()
	lastArticle := min(m.main.articles.offset+m.articleRowsVisible(), len(articles))
	for i := m.main.articles.offset; i < lastArticle; i++ {
		articleRows = append(articleRows, articleRow(articles[i], i == m.main.articles.cursor, articleInnerWidth))
	}
	if len(articleRows) == 0 {
		articleRows = []string{emptyListText}
	}

	feedBox := titledBox(strings.Join(feedRows, "\n"), "Feeds", feedPanelWidth, m.panelHeight(), feedBorderColor)
	articleBox := titledBox(strings.Join(articleRows, "\n"), articleTitle, m.articlePanelWidth(), m.panelHeight(), articleBorderColor)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, feedBox, articleBox)

	status := lipgloss.NewStyle().MaxWidth(m.width).Render(" " + m.mainStatus())

	return lipgloss.JoinVertical(lipgloss.Left, header, panels, status)
}

// formatPubDate formats an item's publication date for the article list,
// falling back to the raw string if it can't be parsed.
func formatPubDate(item rss.Item) string {
	if t, ok := item.Published(); ok {
		return t.Format("02 Jan 2006")
	}
	return truncateString(item.PubDate, 16)
}
