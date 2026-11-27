package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
)

var (
	testProvider = config.Provider{
		Name:        "Test",
		Description: "Test News",
		FullText:    true,
		Feeds: []config.Feed{
			{Name: "One", URL: "https://test.example/one.xml"},
			{Name: "Two", URL: "https://test.example/two.xml"},
		},
	}
	summaryProvider = config.Provider{
		Name:        "Summaries",
		Description: "Summaries only",
		Feeds: []config.Feed{
			{Name: "Headlines", URL: "https://summaries.example/rss"},
		},
	}
)

// fakeFeeds returns n items for each requested feed, or an error for feeds
// named in failing.
type fakeFeeds struct {
	items   map[string]int
	failing map[string]bool
}

func (f fakeFeeds) FetchAll(_ context.Context, requests []rss.FeedRequest) []rss.FeedResult {
	results := make([]rss.FeedResult, len(requests))
	for i, req := range requests {
		results[i].FeedRequest = req
		if f.failing[req.Name] {
			results[i].Err = errors.New("connection refused")
			continue
		}
		for j := range f.items[req.Name] {
			results[i].Feed.Items = append(results[i].Feed.Items, rss.Item{
				Title:       fmt.Sprintf("%s story %d with a headline long enough to need truncating on narrow screens", req.Name, j),
				Link:        fmt.Sprintf("https://test.example/%s/%d", req.Name, j),
				PubDate:     "Sun, 18 Jan 2026 19:25:50 +1300",
				Author:      "A. Reporter",
				Description: fmt.Sprintf("<p>Summary of %s story %d.</p>", req.Name, j),
			})
		}
	}
	return results
}

// fakeArticles returns an article titled after the URL, or err if set. It
// counts calls in fetches if that's non-nil.
type fakeArticles struct {
	err     error
	fetches *int
}

func (f fakeArticles) Fetch(_ context.Context, url string) (*article.Article, error) {
	if f.fetches != nil {
		*f.fetches++
	}
	if f.err != nil {
		return nil, f.err
	}
	return &article.Article{URL: url, Title: "Article at " + url, Content: strings.Repeat("Some article text. ", 200)}, nil
}

func newTestModel(t *testing.T, feeds fakeFeeds, articles fakeArticles, width, height int) Model {
	t.Helper()
	providers := []config.Provider{testProvider, summaryProvider}
	m, _ := send(t, NewModel(providers, feeds, articles), tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func defaultFeeds() fakeFeeds {
	return fakeFeeds{items: map[string]int{"One": 5, "Two": 1}}
}

// send delivers msg and then runs each returned command, feeding its message
// back in as the Bubble Tea runtime would, until there's no command left. It
// reports whether the program quit. Don't use it for keys whose command has
// side effects, like "o".
func send(t *testing.T, m Model, msg tea.Msg) (Model, bool) {
	t.Helper()
	for msg != nil {
		next, cmd := m.Update(msg)
		m = next.(Model)
		if cmd == nil {
			break
		}
		msg = cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			return m, true
		}
	}
	return m, false
}

func press(t *testing.T, m Model, keys ...string) (Model, bool) {
	t.Helper()
	for _, k := range keys {
		var quit bool
		if m, quit = send(t, m, keyMsg(k)); quit {
			return m, true
		}
	}
	return m, false
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func TestSelectProviderLoadsFeeds(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter")

	if m.activeScreen != screenMain {
		t.Fatalf("activeScreen = %v, want main", m.activeScreen)
	}
	if m.currentFeedName() != "One" {
		t.Errorf("currentFeedName = %q, want One", m.currentFeedName())
	}
	if n := len(m.currentArticles()); n != 5 {
		t.Errorf("article list has %d items, want 5", n)
	}
	if m.main.loading {
		t.Error("still loading after feeds arrived")
	}
}

func TestFeedErrorShownAndCleared(t *testing.T) {
	feeds := defaultFeeds()
	feeds.failing = map[string]bool{"One": true}
	m := newTestModel(t, feeds, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter")

	if s := m.mainStatus(); !strings.Contains(s, "Error loading One") {
		t.Errorf("status = %q, want the first feed's error", s)
	}

	m, _ = press(t, m, "j")
	if s := m.mainStatus(); strings.Contains(s, "Error") {
		t.Errorf("status = %q after moving to a healthy feed, want no error", s)
	}
}

func TestEscDoesNotQuit(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	if _, quit := press(t, m, "esc"); quit {
		t.Error("esc on the splash screen quit")
	}
	m, _ = press(t, m, "enter")
	if _, quit := press(t, m, "esc"); quit {
		t.Error("esc on the main screen quit")
	}
}

func TestCtrlCQuitsFromEveryScreen(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	if _, quit := press(t, m, "ctrl+c"); !quit {
		t.Error("ctrl+c didn't quit from splash")
	}
	m, _ = press(t, m, "enter")
	if _, quit := press(t, m, "ctrl+c"); !quit {
		t.Error("ctrl+c didn't quit from main")
	}
	m, _ = press(t, m, "l", "enter")
	if m.activeScreen != screenArticle {
		t.Fatalf("activeScreen = %v, want article", m.activeScreen)
	}
	if _, quit := press(t, m, "ctrl+c"); !quit {
		t.Error("ctrl+c didn't quit from the article view")
	}
	m, _ = press(t, m, "?")
	if _, quit := press(t, m, "ctrl+c"); !quit {
		t.Error("ctrl+c didn't quit from help")
	}
}

func TestOpenArticle(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l", "j", "enter")

	if m.activeScreen != screenArticle {
		t.Fatalf("activeScreen = %v, want article", m.activeScreen)
	}
	want := "https://test.example/One/1"
	if m.article.article == nil || m.article.article.URL != want {
		t.Fatalf("article = %+v, want the one at %s", m.article.article, want)
	}

	m, _ = press(t, m, "q")
	if m.activeScreen != screenMain {
		t.Errorf("activeScreen after q = %v, want main", m.activeScreen)
	}
}

func TestStaleArticleResultIgnored(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter")

	first := rss.Item{Title: "First", Link: "https://test.example/first"}
	second := rss.Item{Title: "Second", Link: "https://test.example/second"}

	firstCmd := m.openArticle(first)
	m, _ = press(t, m, "q")
	secondCmd := m.openArticle(second)

	// The first fetch finishes after the user has moved on to the second.
	m, _ = send(t, m, firstCmd())
	if m.article.article != nil {
		t.Fatalf("stale result replaced the loading view with %q", m.article.article.Title)
	}

	m, _ = send(t, m, secondCmd())
	if m.article.article == nil || m.article.article.URL != second.Link {
		t.Errorf("article = %+v, want %s", m.article.article, second.Link)
	}
}

func TestArticleErrorFallsBackToSummary(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{err: errors.New("no parser")}, 80, 24)
	m, _ = press(t, m, "enter", "l", "enter")

	art := m.article.article
	if art == nil || art.Content != "Summary of One story 0." || art.Author != "A. Reporter" {
		t.Fatalf("article = %+v, want the feed summary", art)
	}
	if !strings.Contains(m.article.note, "no parser") {
		t.Errorf("note = %q, want the error", m.article.note)
	}
	if want := "https://test.example/One/0"; m.article.url != want {
		t.Errorf("url = %q, want %q", m.article.url, want)
	}
	// Check a command is returned without running it, which would open a browser.
	if _, cmd := m.Update(keyMsg("o")); cmd == nil {
		t.Error(`"o" after an error did nothing`)
	}
}

func TestArticleCursorResetsOnFeedChange(t *testing.T) {
	m := newTestModel(t, fakeFeeds{items: map[string]int{"One": 5, "Two": 5}}, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l", "j", "j")
	if i := m.main.articles.cursor; i != 2 {
		t.Fatalf("article index = %d, want 2", i)
	}

	m, _ = press(t, m, "h", "j")
	if m.currentFeedName() != "Two" {
		t.Fatalf("currentFeedName = %q, want Two", m.currentFeedName())
	}
	if i := m.main.articles.cursor; i != 0 {
		t.Errorf("article index after feed change = %d, want 0", i)
	}
}

func TestFeedSelectionKeptWhenFeedsArriveLate(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)

	// Select the provider but hold back the feed fetch.
	cmd := m.selectProvider(testProvider)

	m, _ = press(t, m, "j")
	m, _ = send(t, m, cmd())

	if m.currentFeedName() != "Two" || m.main.feeds.cursor != 1 {
		t.Fatalf("currentFeedName = %q (cursor %d), want Two (cursor 1)", m.currentFeedName(), m.main.feeds.cursor)
	}
	if n := len(m.currentArticles()); n != 1 {
		t.Errorf("article list has %d items, want Two's 1", n)
	}
}

func TestRefresh(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	loadCmd := m.selectProvider(testProvider)

	// A refresh while the first load is in flight is ignored.
	if _, cmd := m.Update(keyMsg("r")); cmd != nil {
		t.Error("refresh started while feeds were loading")
	}

	m, _ = send(t, m, loadCmd())
	m, _ = press(t, m, "r")
	if s := m.mainStatus(); !strings.Contains(s, "Feeds refreshed") {
		t.Errorf("status = %q, want refresh notice", s)
	}

	// Results from a superseded fetch are dropped.
	m, _ = send(t, m, feedsMsg{reqID: m.feedReqID - 1})
	if n := len(m.currentArticles()); n != 5 {
		t.Errorf("stale feed results replaced the article list (%d items)", n)
	}
}

func TestHelpReturnsToPreviousScreen(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l", "enter", "?")
	if m.activeScreen != screenHelp {
		t.Fatalf("activeScreen = %v, want help", m.activeScreen)
	}
	m, _ = press(t, m, "x")
	if m.activeScreen != screenArticle {
		t.Errorf("activeScreen after closing help = %v, want article", m.activeScreen)
	}
}

func TestResizeRewrapsArticle(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 120, 40)
	m, _ = press(t, m, "enter", "l", "enter")
	wide := m.article.viewport.TotalLineCount()

	m, _ = send(t, m, tea.WindowSizeMsg{Width: 60, Height: 40})
	if narrow := m.article.viewport.TotalLineCount(); narrow <= wide {
		t.Errorf("article has %d lines at width 60 and %d at 120; want more when narrower", narrow, wide)
	}
}

func TestViewFillsWindow(t *testing.T) {
	screens := []struct {
		name string
		keys []string
	}{
		{"splash", nil},
		{"main", []string{"enter"}},
		{"article", []string{"enter", "l", "enter"}},
		{"help", []string{"enter", "?"}},
	}
	sizes := []struct{ w, h int }{{80, 24}, {120, 40}, {60, 20}, {40, 12}}

	for _, s := range screens {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%s/%dx%d", s.name, size.w, size.h), func(t *testing.T) {
				m := newTestModel(t, defaultFeeds(), fakeArticles{}, size.w, size.h)
				m, _ = press(t, m, s.keys...)

				lines := strings.Split(m.View(), "\n")
				if len(lines) != size.h {
					t.Errorf("view has %d lines, want %d", len(lines), size.h)
				}
				for i, line := range lines {
					if w := lipgloss.Width(line); w > size.w {
						t.Errorf("line %d is %d wide, want <= %d: %q", i, w, size.w, line)
					}
				}
			})
		}
	}
}

func TestSummaryProviderDoesNotFetchArticles(t *testing.T) {
	var fetches int
	m := newTestModel(t, fakeFeeds{items: map[string]int{"Headlines": 2}}, fakeArticles{fetches: &fetches}, 80, 24)
	m, _ = press(t, m, "j", "enter", "l", "enter")

	if m.activeScreen != screenArticle {
		t.Fatalf("activeScreen = %v, want article", m.activeScreen)
	}
	if fetches != 0 {
		t.Errorf("fetched the article page %d times, want 0", fetches)
	}
	art := m.article.article
	if art == nil || art.Content != "Summary of Headlines story 0." || art.Source != "Summaries" {
		t.Fatalf("article = %+v, want the feed summary", art)
	}
	if m.article.note == "" || strings.Contains(m.article.note, "Couldn't") {
		t.Errorf("note = %q, want the summary notice without an error", m.article.note)
	}
}

func TestBackToProviders(t *testing.T) {
	m := newTestModel(t, fakeFeeds{items: map[string]int{"One": 5, "Two": 5, "Headlines": 2}}, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "j", "l", "j")

	for _, k := range []string{"esc", "backspace"} {
		m, _ = press(t, m, k)
		if m.activeScreen != screenSplash {
			t.Fatalf("activeScreen after %s = %v, want splash", k, m.activeScreen)
		}
		m, _ = press(t, m, "enter")
	}

	// Switch to the other provider: nothing carries over.
	m, _ = press(t, m, "esc", "j", "enter")
	if m.provider.Name != "Summaries" || m.currentFeedName() != "Headlines" {
		t.Fatalf("provider/feed = %q/%q, want Summaries/Headlines", m.provider.Name, m.currentFeedName())
	}
	if m.main.focus != focusFeedList || m.main.feeds.cursor != 0 || m.main.articles.cursor != 0 {
		t.Errorf("focus/cursors not reset: focus=%v feed=%d article=%d",
			m.main.focus, m.main.feeds.cursor, m.main.articles.cursor)
	}
	if _, ok := m.main.feedResults["One"]; ok {
		t.Error("previous provider's feed results are still stored")
	}
	if !strings.Contains(m.View(), "Knews - Summaries") {
		t.Error("header doesn't name the provider")
	}
}

func TestFeedsFromPreviousProviderIgnored(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)

	// Pick the first provider, then leave before its feeds arrive.
	staleCmd := m.selectProvider(testProvider)
	m, _ = press(t, m, "esc", "j", "enter")
	if m.provider.Name != "Summaries" {
		t.Fatalf("provider = %q, want Summaries", m.provider.Name)
	}

	m, _ = send(t, m, staleCmd())
	if _, ok := m.main.feedResults["One"]; ok {
		t.Error("feeds for the previous provider were stored")
	}
}

func TestMainListScrolling(t *testing.T) {
	m := newTestModel(t, fakeFeeds{items: map[string]int{"One": 30}}, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l")

	// Each article takes two lines, so far fewer than 30 are visible at once.
	visible := (m.panelHeight() - 2) / 2
	for range 20 {
		m, _ = press(t, m, "j")
		cursor, offset := m.main.articles.cursor, m.main.articles.offset
		if cursor < offset || cursor >= offset+visible {
			t.Fatalf("cursor %d is outside the visible rows %d..%d", cursor, offset, offset+visible-1)
		}
	}
	if m.main.articles.cursor != 20 || m.main.articles.offset == 0 {
		t.Errorf("cursor = %d, offset = %d; want cursor 20 and a scrolled list", m.main.articles.cursor, m.main.articles.offset)
	}
	if !strings.Contains(m.View(), "One story 20") {
		t.Error("the selected article isn't drawn")
	}
	if strings.Contains(m.View(), "One story 0 ") {
		t.Error("the first article is still drawn after scrolling past it")
	}

	// The cursor stops at the last article.
	for range 20 {
		m, _ = press(t, m, "j")
	}
	if m.main.articles.cursor != 29 {
		t.Errorf("cursor = %d, want 29 (the last article)", m.main.articles.cursor)
	}
}

func TestListNavigationKeys(t *testing.T) {
	m := newTestModel(t, fakeFeeds{items: map[string]int{"One": 30, "Two": 30}}, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l")
	page := m.articleRowsVisible()

	// The article panel: top, bottom and a page at a time.
	steps := []struct {
		key  string
		want int
	}{
		{"G", 29}, {"g", 0}, {"end", 29}, {"home", 0},
		{"d", page}, {"u", 0}, {"f", page}, {"b", 0},
		{"pgdown", page}, {"pgup", 0},
	}
	for _, step := range steps {
		m, _ = press(t, m, step.key)
		if got := m.main.articles.cursor; got != step.want {
			t.Errorf("after %q article cursor = %d, want %d", step.key, got, step.want)
		}
	}

	// The feed panel: G moves to the last feed and shows its articles from
	// the top.
	m, _ = press(t, m, "j", "j", "h", "G")
	if m.currentFeedName() != "Two" {
		t.Errorf("after G on the feed panel, feed = %q, want Two", m.currentFeedName())
	}
	if m.main.articles.cursor != 0 {
		t.Errorf("article cursor after changing feed = %d, want 0", m.main.articles.cursor)
	}

	// The provider list: G then Enter picks the last provider.
	m, _ = press(t, m, "esc", "G", "enter")
	if m.provider.Name != "Summaries" {
		t.Errorf("after G and Enter on the provider list, provider = %q, want Summaries", m.provider.Name)
	}
}

func TestEmptyListShowsNoItems(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)

	// Select the provider but hold back the feed fetch.
	cmd := m.selectProvider(testProvider)
	if !strings.Contains(m.View(), "No items.") {
		t.Error("the article panel isn't marked empty while feeds load")
	}

	m, _ = send(t, m, cmd())
	if strings.Contains(m.View(), "No items.") {
		t.Error("the article panel is still marked empty after feeds arrived")
	}
}

func TestHelpClearsNotice(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "r")
	if !strings.Contains(m.mainStatus(), "Feeds refreshed") {
		t.Fatalf("status = %q, want refresh notice", m.mainStatus())
	}

	m, _ = press(t, m, "?", "x")
	if s := m.mainStatus(); strings.Contains(s, "Feeds refreshed") {
		t.Errorf("status = %q after opening and closing help, want the notice cleared", s)
	}
}
