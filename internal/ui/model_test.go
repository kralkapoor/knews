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

var testProvider = config.Provider{
	Name:        "Test",
	Description: "Test News",
	Feeds: []config.Feed{
		{Name: "One", URL: "https://test.example/one.xml"},
		{Name: "Two", URL: "https://test.example/two.xml"},
	},
}

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
			results[i].Feed.Channel.Items = append(results[i].Feed.Channel.Items, rss.Item{
				Title:   fmt.Sprintf("%s story %d with a headline long enough to need truncating on narrow screens", req.Name, j),
				Link:    fmt.Sprintf("https://test.example/%s/%d", req.Name, j),
				PubDate: "Sun, 18 Jan 2026 19:25:50 +1300",
			})
		}
	}
	return results
}

// fakeArticles returns an article titled after the URL, or err if set.
type fakeArticles struct {
	err error
}

func (f fakeArticles) Fetch(_ context.Context, url string) (*article.Article, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &article.Article{URL: url, Title: "Article at " + url, Content: strings.Repeat("Some article text. ", 200)}, nil
}

func newTestModel(t *testing.T, feeds fakeFeeds, articles fakeArticles, width, height int) Model {
	t.Helper()
	m, _ := send(t, NewModel([]config.Provider{testProvider}, feeds, articles), tea.WindowSizeMsg{Width: width, Height: height})
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
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func TestSelectProviderLoadsFeeds(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter")

	if m.activeScreen != screenMain {
		t.Fatalf("activeScreen = %v, want main", m.activeScreen)
	}
	if m.main.currentFeed != "One" {
		t.Errorf("currentFeed = %q, want One", m.main.currentFeed)
	}
	if n := len(m.main.articleList.Items()); n != 5 {
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

	if s := m.main.status(); !strings.Contains(s, "Error loading One") {
		t.Errorf("status = %q, want the first feed's error", s)
	}

	m, _ = press(t, m, "j")
	if s := m.main.status(); strings.Contains(s, "Error") {
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
	if m.articleView.article == nil || m.articleView.article.URL != want {
		t.Fatalf("article = %+v, want the one at %s", m.articleView.article, want)
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

	next, firstCmd := m.Update(openArticleMsg{item: first})
	m, _ = send(t, next.(Model), closeArticleMsg{})
	next, secondCmd := m.Update(openArticleMsg{item: second})
	m = next.(Model)

	// The first fetch finishes after the user has moved on to the second.
	m, _ = send(t, m, firstCmd())
	if m.articleView.article != nil {
		t.Fatalf("stale result replaced the loading view with %q", m.articleView.article.Title)
	}

	m, _ = send(t, m, secondCmd())
	if m.articleView.article == nil || m.articleView.article.URL != second.Link {
		t.Errorf("article = %+v, want %s", m.articleView.article, second.Link)
	}
}

func TestArticleErrorKeepsOpenInBrowser(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{err: errors.New("no parser")}, 80, 24)
	m, _ = press(t, m, "enter", "l", "enter")

	if !strings.Contains(m.articleView.message, "no parser") {
		t.Errorf("message = %q, want the error", m.articleView.message)
	}
	if want := "https://test.example/One/0"; m.articleView.url != want {
		t.Errorf("url = %q, want %q", m.articleView.url, want)
	}
	// Check a command is returned without running it, which would open a browser.
	if _, cmd := m.Update(keyMsg("o")); cmd == nil {
		t.Error(`"o" after an error did nothing`)
	}
}

func TestArticleCursorResetsOnFeedChange(t *testing.T) {
	m := newTestModel(t, fakeFeeds{items: map[string]int{"One": 5, "Two": 5}}, fakeArticles{}, 80, 24)
	m, _ = press(t, m, "enter", "l", "j", "j")
	if i := m.main.articleList.Index(); i != 2 {
		t.Fatalf("article index = %d, want 2", i)
	}

	m, _ = press(t, m, "h", "j")
	if m.main.currentFeed != "Two" {
		t.Fatalf("currentFeed = %q, want Two", m.main.currentFeed)
	}
	if i := m.main.articleList.Index(); i != 0 {
		t.Errorf("article index after feed change = %d, want 0", i)
	}
}

func TestFeedSelectionKeptWhenFeedsArriveLate(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)

	// Select the provider but hold back the feed fetch.
	next, cmd := m.Update(providerSelectedMsg{provider: testProvider})
	m = next.(Model)

	m, _ = press(t, m, "j")
	m, _ = send(t, m, cmd())

	if m.main.currentFeed != "Two" || m.main.feedList.Index() != 1 {
		t.Fatalf("currentFeed = %q (cursor %d), want Two (cursor 1)", m.main.currentFeed, m.main.feedList.Index())
	}
	if n := len(m.main.articleList.Items()); n != 1 {
		t.Errorf("article list has %d items, want Two's 1", n)
	}
}

func TestRefresh(t *testing.T) {
	m := newTestModel(t, defaultFeeds(), fakeArticles{}, 80, 24)
	next, loadCmd := m.Update(providerSelectedMsg{provider: testProvider})
	m = next.(Model)

	// A refresh while the first load is in flight is ignored.
	if _, cmd := m.Update(requestRefreshMsg{}); cmd != nil {
		t.Error("refresh started while feeds were loading")
	}

	m, _ = send(t, m, loadCmd())
	m, _ = press(t, m, "r")
	if s := m.main.status(); !strings.Contains(s, "Feeds refreshed") {
		t.Errorf("status = %q, want refresh notice", s)
	}

	// Results from a superseded fetch are dropped.
	m, _ = send(t, m, feedsMsg{reqID: m.feedReqID - 1})
	if n := len(m.main.articleList.Items()); n != 5 {
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
	wide := m.articleView.viewport.TotalLineCount()

	m, _ = send(t, m, tea.WindowSizeMsg{Width: 60, Height: 40})
	if narrow := m.articleView.viewport.TotalLineCount(); narrow <= wide {
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
