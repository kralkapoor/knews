package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
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

const feedPanelWidth = 22 // 20 inner + 2 for borders

// feedItem implements list.Item for the feed category list.
type feedItem struct {
	name string
}

func (f feedItem) Title() string       { return f.name }
func (f feedItem) Description() string { return "" }
func (f feedItem) FilterValue() string { return f.name }

// feedDelegate renders items in the feed list (single-line, no description).
type feedDelegate struct {
	width int
}

func (d feedDelegate) Height() int                             { return 1 }
func (d feedDelegate) Spacing() int                            { return 0 }
func (d feedDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d feedDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	fi, ok := item.(feedItem)
	if !ok {
		return
	}

	style := lipgloss.NewStyle().PaddingLeft(1)
	if index == m.Index() {
		style = style.Background(colorDarkCyan).Foreground(lipgloss.Color("15")).Bold(true)
	}

	name := style.Width(d.width).Render(fi.name)
	fmt.Fprint(w, name)
}

// articleItem implements list.Item for the article list.
type articleItem struct {
	title   string
	dateStr string
	item    rss.Item
}

func (a articleItem) Title() string       { return a.title }
func (a articleItem) Description() string { return a.dateStr }
func (a articleItem) FilterValue() string { return a.title }

// articleDelegate renders items in the article list (title + date).
type articleDelegate struct {
	width int
}

func (d articleDelegate) Height() int                             { return 2 }
func (d articleDelegate) Spacing() int                            { return 0 }
func (d articleDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d articleDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ai, ok := item.(articleItem)
	if !ok {
		return
	}

	selected := index == m.Index()

	titleStyle := lipgloss.NewStyle().PaddingLeft(1)
	dateStyle := lipgloss.NewStyle().PaddingLeft(3).Foreground(colorGray)

	if selected {
		titleStyle = titleStyle.Background(colorDarkCyan).Foreground(lipgloss.Color("15")).Bold(true)
		dateStyle = dateStyle.Background(colorDarkCyan).Foreground(lipgloss.Color("15"))
	}

	title := titleStyle.Width(d.width).Render(ai.title)
	date := dateStyle.Width(d.width).Render(ai.dateStr)
	fmt.Fprintf(w, "%s\n%s", title, date)
}

type mainModel struct {
	feedList    list.Model
	articleList list.Model
	focus       focusPanel
	feedResults map[string]rss.FeedResult
	currentFeed string
	statusMsg   string
	width       int
	height      int
}

func newMainModel() mainModel {
	fl := list.New(nil, feedDelegate{width: feedPanelWidth - 2}, feedPanelWidth-2, 10)
	fl.SetShowTitle(false)
	fl.SetShowStatusBar(false)
	fl.SetShowFilter(false)
	fl.SetShowHelp(false)
	fl.SetShowPagination(false)

	al := list.New(nil, articleDelegate{width: 60}, 60, 10)
	al.SetShowTitle(false)
	al.SetShowStatusBar(false)
	al.SetShowFilter(false)
	al.SetShowHelp(false)
	al.SetShowPagination(false)

	return mainModel{
		feedList:    fl,
		articleList: al,
		focus:       focusFeedList,
		feedResults: make(map[string]rss.FeedResult),
		statusMsg:   statusText("Press ", "?", " for help │ ", "q", " to quit"),
	}
}

func (m *mainModel) setSize(width, height int) {
	m.width = width
	m.height = height

	// Header: 3 lines, status bar: 1 line, borders: 2 lines
	listHeight := height - 3 - 1 - 2
	if listHeight < 1 {
		listHeight = 1
	}

	// Feed list: fixed width
	feedInner := feedPanelWidth - 2
	m.feedList.SetSize(feedInner, listHeight)
	m.feedList.SetDelegate(feedDelegate{width: feedInner})

	// Article list: remaining width minus feed panel, minus border chars
	articleWidth := width - feedPanelWidth - 2
	if articleWidth < 10 {
		articleWidth = 10
	}
	m.articleList.SetSize(articleWidth, listHeight)
	m.articleList.SetDelegate(articleDelegate{width: articleWidth})
}

// setFeeds populates the feed list with category names.
func (m *mainModel) setFeeds(feedNames []string) {
	items := make([]list.Item, len(feedNames))
	for i, name := range feedNames {
		items[i] = feedItem{name: name}
	}
	m.feedList.SetItems(items)
}

// setFeedResults stores fetched feed data and selects the given feed.
func (m *mainModel) setFeedResults(results []rss.FeedResult, selectFeed string) {
	for _, result := range results {
		m.feedResults[result.FeedName] = result
	}
	if selectFeed != "" {
		m.selectFeed(selectFeed)
	}
}

// selectFeed updates the article list for the given feed name.
func (m *mainModel) selectFeed(feedName string) {
	m.currentFeed = feedName

	result, exists := m.feedResults[feedName]
	if !exists || result.Error != nil {
		m.articleList.SetItems(nil)
		if exists && result.Error != nil {
			m.statusMsg = errorStyle.Render("Error loading "+feedName+": ") + result.Error.Error()
		}
		return
	}

	items := make([]list.Item, len(result.Feed.Channel.Items))
	for i, item := range result.Feed.Channel.Items {
		items[i] = articleItem{
			title:   truncateString(item.Title, 60),
			dateStr: formatPubDate(item.PubDate),
			item:    item,
		}
	}
	m.articleList.SetItems(items)
}

func (m mainModel) Update(msg tea.Msg) (mainModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.toggleFocus()
			return m, nil
		case "h", "left":
			m.focus = focusFeedList
			return m, nil
		case "l", "right":
			m.focus = focusArticleList
			return m, nil
		case "q":
			return m, tea.Quit
		case "?":
			return m, func() tea.Msg { return showHelpMsg{} }
		case "r":
			return m, func() tea.Msg { return requestRefreshMsg{} }
		case "enter":
			if m.focus == focusFeedList {
				// Select feed and switch focus to article list
				m.focus = focusArticleList
				return m, nil
			}
			// Open selected article
			if item, ok := m.articleList.SelectedItem().(articleItem); ok {
				return m, func() tea.Msg { return openArticleMsg{item: item.item} }
			}
			return m, nil
		}

		// Forward j/k/up/down to the focused list
		prevFeedIdx := m.feedList.Index()
		if m.focus == focusFeedList {
			var cmd tea.Cmd
			m.feedList, cmd = m.feedList.Update(msg)
			// Detect cursor change to update article list
			if m.feedList.Index() != prevFeedIdx {
				if fi, ok := m.feedList.SelectedItem().(feedItem); ok {
					m.selectFeed(fi.name)
				}
			}
			return m, cmd
		}
		var cmd tea.Cmd
		m.articleList, cmd = m.articleList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m mainModel) View() string {
	// Header
	headerText := titleStyle.Render("Knews - NZ News Reader")
	headerContent := lipgloss.PlaceHorizontal(m.width-2, lipgloss.Center, headerText)
	header := titledBox(headerContent, "", m.width, 3, colorDarkCyan)

	// Feed list panel
	feedBorderColor := colorGray
	if m.focus == focusFeedList {
		feedBorderColor = colorDarkCyan
	}
	feedBox := titledBox(m.feedList.View(), "Feeds", feedPanelWidth, m.height-3-1, feedBorderColor)

	// Article list panel
	articleBorderColor := colorGray
	if m.focus == focusArticleList {
		articleBorderColor = colorDarkCyan
	}
	articleTitle := "Articles"
	if m.currentFeed != "" {
		articleTitle = m.currentFeed
	}
	articleWidth := m.width - feedPanelWidth
	articleBox := titledBox(m.articleList.View(), articleTitle, articleWidth, m.height-3-1, articleBorderColor)

	// Two-column layout
	panels := lipgloss.JoinHorizontal(lipgloss.Top, feedBox, articleBox)

	// Status bar
	status := " " + m.statusMsg

	return lipgloss.JoinVertical(lipgloss.Left, header, panels, status)
}

func (m *mainModel) toggleFocus() {
	if m.focus == focusFeedList {
		m.focus = focusArticleList
	} else {
		m.focus = focusFeedList
	}
}

// truncateString shortens a string to maxLength runes, adding an ellipsis if
// needed. It counts runes rather than bytes so multi-byte characters (e.g. the
// macrons in te reo Māori) are never split mid-character.
func truncateString(s string, maxLength int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	if maxLength < 3 {
		return string(runes[:maxLength])
	}
	return string(runes[:maxLength-3]) + "..."
}

// formatPubDate parses and formats an RSS publication date string.
func formatPubDate(pubDate string) string {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, pubDate); err == nil {
			return parsedTime.Format("02 Jan 2006")
		}
	}

	if runes := []rune(pubDate); len(runes) > 16 {
		return string(runes[:16])
	}
	return pubDate
}
