package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/rivo/tview"
)

// ArticleList represents the article list panel showing items from the selected feed.
type ArticleList struct {
	list         *tview.List
	app          *App
	currentItems []rss.Item
}

// NewArticleList creates a new ArticleList panel.
func NewArticleList(app *App) *ArticleList {
	al := &ArticleList{
		list:         tview.NewList(),
		app:          app,
		currentItems: make([]rss.Item, 0),
	}

	al.list.SetBorder(true).
		SetTitle(" Articles ").
		SetBorderColor(tcell.ColorGray)

	al.list.ShowSecondaryText(true)
	al.list.SetHighlightFullLine(true)
	al.list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	al.list.SetSelectedTextColor(tcell.ColorWhite)
	al.list.SetSecondaryTextColor(tcell.ColorGray)

	// Set up selection handler
	al.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(al.currentItems) {
			app.openArticle(al.currentItems[index])
		}
	})

	// Handle j/k navigation
	al.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				currentIndex := al.list.GetCurrentItem()
				if currentIndex < al.list.GetItemCount()-1 {
					al.list.SetCurrentItem(currentIndex + 1)
				}
				return nil
			case 'k':
				currentIndex := al.list.GetCurrentItem()
				if currentIndex > 0 {
					al.list.SetCurrentItem(currentIndex - 1)
				}
				return nil
			}
		}
		return event
	})

	return al
}

// SetArticles updates the article list with items from a feed.
func (al *ArticleList) SetArticles(items []rss.Item, feedName string) {
	al.list.Clear()
	al.currentItems = items
	al.list.SetTitle(" " + feedName + " ")

	for _, item := range items {
		title := truncateString(item.Title, 60)
		dateStr := formatPubDate(item.PubDate)
		al.list.AddItem(title, dateStr, 0, nil)
	}

	// Select first item if available
	if len(items) > 0 {
		al.list.SetCurrentItem(0)
	}
}

// Clear removes all articles from the list.
func (al *ArticleList) Clear() {
	al.list.Clear()
	al.currentItems = nil
	al.list.SetTitle(" Articles ")
}

// GetSelectedItem returns the currently selected article item.
func (al *ArticleList) GetSelectedItem() *rss.Item {
	currentIndex := al.list.GetCurrentItem()
	if currentIndex >= 0 && currentIndex < len(al.currentItems) {
		return &al.currentItems[currentIndex]
	}
	return nil
}

// truncateString shortens a string to the specified length, adding ellipsis if needed.
func truncateString(s string, maxLength int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// formatPubDate parses and formats an RSS publication date string.
func formatPubDate(pubDate string) string {
	// Try common RSS date formats
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

	// If parsing fails, return original string truncated
	if len(pubDate) > 16 {
		return pubDate[:16]
	}
	return pubDate
}
