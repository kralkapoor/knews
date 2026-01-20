package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FeedList represents the feed category selection panel.
type FeedList struct {
	list *tview.List
	app  *App
}

// NewFeedList creates a new FeedList panel.
func NewFeedList(app *App) *FeedList {
	fl := &FeedList{
		list: tview.NewList(),
		app:  app,
	}

	fl.list.SetBorder(true).
		SetTitle(" Feeds ").
		SetBorderColor(tcell.ColorDarkCyan)

	fl.list.ShowSecondaryText(false)
	fl.list.SetHighlightFullLine(true)
	fl.list.SetSelectedBackgroundColor(tcell.ColorDarkCyan)
	fl.list.SetSelectedTextColor(tcell.ColorWhite)

	// Set up selection handler
	fl.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		app.selectFeed(mainText)
		app.setFocus(FocusArticleList)
	})

	// Handle j/k navigation
	fl.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				currentIndex := fl.list.GetCurrentItem()
				if currentIndex < fl.list.GetItemCount()-1 {
					fl.list.SetCurrentItem(currentIndex + 1)
				}
				return nil
			case 'k':
				currentIndex := fl.list.GetCurrentItem()
				if currentIndex > 0 {
					fl.list.SetCurrentItem(currentIndex - 1)
				}
				return nil
			}
		}
		return event
	})

	// Handle changed event to update article list
	fl.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		app.selectFeed(mainText)
	})

	return fl
}

// addFeed adds a feed category to the list.
func (fl *FeedList) addFeed(name string) {
	fl.list.AddItem(name, "", 0, nil)
}

// SetFeeds replaces the current feed list with the provided feed names.
func (fl *FeedList) SetFeeds(feedNames []string) {
	fl.list.Clear()
	for _, feedName := range feedNames {
		fl.addFeed(feedName)
	}
}

// GetCurrentFeed returns the currently selected feed name.
func (fl *FeedList) GetCurrentFeed() string {
	currentIndex := fl.list.GetCurrentItem()
	mainText, _ := fl.list.GetItemText(currentIndex)
	return mainText
}
