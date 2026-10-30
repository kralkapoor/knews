package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/config"
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

// newList returns a list.Model with the built-in title, filter, status bar,
// help, pagination and quit keys turned off. The screens draw their own
// chrome and handle quitting themselves.
func newList(delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(nil, delegate, width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.DisableQuitKeybindings()
	return l
}

// feedItem implements list.Item for the feed category list.
type feedItem struct {
	name string
}

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

	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = selectedStyle.Bold(true)
	}
	fmt.Fprint(w, style.PaddingLeft(1).Width(d.width).Render(truncateString(fi.name, d.width-1)))
}

// articleItem implements list.Item for the article list.
type articleItem struct {
	item rss.Item
}

func (a articleItem) FilterValue() string { return a.item.Title }

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
	base := lipgloss.NewStyle()
	if selected {
		base = selectedStyle
	}
	titleStyle := base.Bold(selected).PaddingLeft(1)
	dateStyle := base.PaddingLeft(3)
	if !selected {
		dateStyle = dateStyle.Foreground(colorGray)
	}

	// Truncate rather than letting lipgloss wrap: a wrapped title would make
	// the item taller than Height() and push the list out of its box.
	title := titleStyle.Width(d.width).Render(truncateString(ai.item.Title, d.width-1))
	date := dateStyle.Width(d.width).Render(truncateString(formatPubDate(ai.item), d.width-3))
	fmt.Fprintf(w, "%s\n%s", title, date)
}

type mainModel struct {
	feedList    list.Model
	articleList list.Model
	focus       focusPanel
	feedResults map[string]rss.FeedResult
	currentFeed string
	loading     bool   // a feed fetch is in flight
	notice      string // transient status message, cleared on the next key
	width       int
	height      int
}

func newMainModel() mainModel {
	return mainModel{
		feedList:    newList(feedDelegate{width: feedPanelWidth - 2}, feedPanelWidth-2, 10),
		articleList: newList(articleDelegate{width: 60}, 60, 10),
		focus:       focusFeedList,
		feedResults: make(map[string]rss.FeedResult),
	}
}

// articleWidth returns the width of the article panel, borders included.
func (m mainModel) articleWidth() int {
	return max(m.width-feedPanelWidth, minArticleWidth+2)
}

// panelHeight returns the height of the two list panels, borders included.
func (m mainModel) panelHeight() int {
	return max(m.height-headerHeight-statusHeight, 3)
}

func (m *mainModel) setSize(width, height int) {
	m.width = width
	m.height = height

	listHeight := m.panelHeight() - 2

	feedInner := feedPanelWidth - 2
	m.feedList.SetSize(feedInner, listHeight)
	m.feedList.SetDelegate(feedDelegate{width: feedInner})

	articleInner := m.articleWidth() - 2
	m.articleList.SetSize(articleInner, listHeight)
	m.articleList.SetDelegate(articleDelegate{width: articleInner})
}

// setFeeds resets the view for a newly selected provider and lists its feeds.
// The first feed is selected; its articles appear once setFeedResults is
// called.
func (m *mainModel) setFeeds(feeds []config.Feed) {
	items := make([]list.Item, len(feeds))
	for i, f := range feeds {
		items[i] = feedItem{name: f.Name}
	}
	m.feedList.SetItems(items)
	m.feedList.ResetSelected()
	m.feedResults = make(map[string]rss.FeedResult, len(feeds))
	m.focus = focusFeedList
	m.notice = ""
	m.currentFeed = ""
	if len(feeds) > 0 {
		m.selectFeed(feeds[0].Name)
	}
}

// setFeedResults stores fetched feed data and refreshes the current feed's
// articles.
func (m *mainModel) setFeedResults(results []rss.FeedResult) {
	for _, result := range results {
		m.feedResults[result.Name] = result
	}
	m.selectFeed(m.currentFeed)
}

// selectFeed shows the articles for the named feed. The article cursor moves
// to the top when the feed changes, and otherwise stays put.
func (m *mainModel) selectFeed(name string) {
	result := m.feedResults[name]
	items := make([]list.Item, len(result.Feed.Channel.Items))
	for i, item := range result.Feed.Channel.Items {
		items[i] = articleItem{item: item}
	}
	m.articleList.SetItems(items)
	if name != m.currentFeed || m.articleList.Index() >= len(items) {
		m.articleList.ResetSelected()
	}
	m.currentFeed = name
}

// status returns the text for the status bar.
func (m mainModel) status() string {
	if m.loading {
		return keyStyle.Render("Loading feeds...")
	}
	if err := m.feedResults[m.currentFeed].Err; err != nil {
		return errorStyle.Render("Error loading "+m.currentFeed+": ") + err.Error()
	}
	hint := statusText("Press ", "?", " for help │ ", "q", " to quit")
	if m.notice != "" {
		return m.notice + " │ " + hint
	}
	return hint
}

func (m mainModel) Update(msg tea.Msg) (mainModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	m.notice = ""

	switch keyMsg.String() {
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
			m.focus = focusArticleList
			return m, nil
		}
		if item, ok := m.articleList.SelectedItem().(articleItem); ok {
			return m, func() tea.Msg { return openArticleMsg{item: item.item} }
		}
		return m, nil
	}

	// Forward everything else (j/k/up/down/...) to the focused list.
	var cmd tea.Cmd
	if m.focus == focusArticleList {
		m.articleList, cmd = m.articleList.Update(msg)
		return m, cmd
	}
	prev := m.feedList.Index()
	m.feedList, cmd = m.feedList.Update(msg)
	if m.feedList.Index() != prev {
		if fi, ok := m.feedList.SelectedItem().(feedItem); ok {
			m.selectFeed(fi.name)
		}
	}
	return m, cmd
}

func (m mainModel) View() string {
	headerText := titleStyle.Render("Knews - NZ News Reader")
	headerContent := lipgloss.PlaceHorizontal(m.width-2, lipgloss.Center, headerText)
	header := titledBox(headerContent, "", m.width, headerHeight, colorDarkCyan)

	feedBorderColor, articleBorderColor := colorGray, colorGray
	if m.focus == focusFeedList {
		feedBorderColor = colorDarkCyan
	} else {
		articleBorderColor = colorDarkCyan
	}

	articleTitle := "Articles"
	if m.currentFeed != "" {
		articleTitle = m.currentFeed
	}

	feedBox := titledBox(m.feedList.View(), "Feeds", feedPanelWidth, m.panelHeight(), feedBorderColor)
	articleBox := titledBox(m.articleList.View(), articleTitle, m.articleWidth(), m.panelHeight(), articleBorderColor)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, feedBox, articleBox)

	status := lipgloss.NewStyle().MaxWidth(m.width).Render(" " + m.status())

	return lipgloss.JoinVertical(lipgloss.Left, header, panels, status)
}

func (m *mainModel) toggleFocus() {
	if m.focus == focusFeedList {
		m.focus = focusArticleList
	} else {
		m.focus = focusFeedList
	}
}

// formatPubDate formats an item's publication date for the article list,
// falling back to the raw string if it can't be parsed.
func formatPubDate(item rss.Item) string {
	if t, ok := item.Published(); ok {
		return t.Format("02 Jan 2006")
	}
	return truncateString(item.PubDate, 16)
}
