package ui

import (
	"context"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/rivo/tview"
)

// FocusPanel represents which panel currently has focus.
type FocusPanel int

const (
	FocusFeedList FocusPanel = iota
	FocusArticleList
)

// App represents the main application state and UI components.
type App struct {
	tviewApp         *tview.Application
	pages            *tview.Pages
	mainFlex         *tview.Flex
	splash           *SplashScreen
	feedList         *FeedList
	articleList      *ArticleList
	articleView      *ArticleView
	statusBar        *tview.TextView
	feedFetcher      *rss.FeedFetcher
	articleFetcher   *article.ArticleFetcher
	feedResults      map[string]rss.FeedResult
	feedResultsMux   sync.RWMutex
	currentFeed      string
	selectedProvider string
	focusPanel       FocusPanel
}

// NewApp creates a new App instance with all UI components initialized.
func NewApp(feedFetcher *rss.FeedFetcher, articleFetcher *article.ArticleFetcher) *App {
	app := &App{
		tviewApp:       tview.NewApplication(),
		pages:          tview.NewPages(),
		feedFetcher:    feedFetcher,
		articleFetcher: articleFetcher,
		feedResults:    make(map[string]rss.FeedResult),
		focusPanel:     FocusFeedList,
	}

	app.setupUI()
	app.setupKeybindings()

	return app
}

// setupUI initializes all UI components and layouts.
func (a *App) setupUI() {
	// Create status bar
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.statusBar.SetBorder(false)
	a.setStatus("Press [yellow]?[white] for help | [yellow]q[white] to quit")

	// Create feed list
	a.feedList = NewFeedList(a)

	// Create article list
	a.articleList = NewArticleList(a)

	// Create article view
	a.articleView = NewArticleView(a)

	// Create left panel (feed list)
	leftPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.feedList.list, 0, 1, true)

	// Create right panel (article list)
	rightPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.articleList.list, 0, 1, false)

	// Create main layout with two columns
	a.mainFlex = tview.NewFlex().
		AddItem(leftPanel, 20, 0, true).
		AddItem(rightPanel, 0, 1, false)

	// Create main page layout with status bar
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.createHeader(), 3, 0, false).
		AddItem(a.mainFlex, 0, 1, true).
		AddItem(a.statusBar, 1, 0, false)

	// Create splash screen
	a.splash = NewSplashScreen(a)

	// Add pages
	a.pages.AddPage("splash", a.splash.layout, true, true)
	a.pages.AddPage("main", mainLayout, true, false)
	a.pages.AddPage("article", a.articleView.layout, true, false)
	a.pages.AddPage("help", a.createHelpPage(), true, false)

	a.tviewApp.SetRoot(a.pages, true)
}

// createHeader creates the header bar with app title.
func (a *App) createHeader() *tview.TextView {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::b]Knews - NZ News Reader[::-]")
	header.SetBorder(true).SetBorderColor(tcell.ColorDarkCyan)
	return header
}

// createHelpPage creates the help page showing keybindings.
func (a *App) createHelpPage() *tview.Flex {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[::b]Keyboard Shortcuts[::-]

[yellow]Navigation:[white]
  j/↓     Move down
  k/↑     Move up
  h/←     Focus feed list
  l/→     Focus article list
  Tab     Switch panels
  Enter   Select item / Open article

[yellow]Actions:[white]
  r       Refresh feeds
  o       Open article in browser
  q       Quit / Back

[yellow]Press any key to close this help[white]`)
	helpText.SetBorder(true).SetTitle(" Help ").SetBorderColor(tcell.ColorYellow)

	helpLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(helpText, 50, 0, true).
			AddItem(nil, 0, 1, false), 15, 0, true).
		AddItem(nil, 0, 1, false)

	return helpLayout
}

// setupKeybindings configures global keyboard handlers.
func (a *App) setupKeybindings() {
	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := a.pages.GetFrontPage()

		// Handle splash page
		if currentPage == "splash" {
			return a.splash.HandleInput(event)
		}

		// Handle help page
		if currentPage == "help" {
			a.pages.SwitchToPage("main")
			a.updateFocus()
			return nil
		}

		// Handle article view
		if currentPage == "article" {
			return a.articleView.HandleInput(event)
		}

		// Main page keybindings
		switch event.Key() {
		case tcell.KeyTab:
			a.toggleFocus()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				a.tviewApp.Stop()
				return nil
			case '?':
				a.pages.SwitchToPage("help")
				return nil
			case 'r':
				a.RefreshFeeds()
				return nil
			case 'h':
				a.setFocus(FocusFeedList)
				return nil
			case 'l':
				a.setFocus(FocusArticleList)
				return nil
			}
		case tcell.KeyLeft:
			a.setFocus(FocusFeedList)
			return nil
		case tcell.KeyRight:
			a.setFocus(FocusArticleList)
			return nil
		}

		return event
	})
}

// toggleFocus switches focus between the two main panels.
func (a *App) toggleFocus() {
	if a.focusPanel == FocusFeedList {
		a.setFocus(FocusArticleList)
	} else {
		a.setFocus(FocusFeedList)
	}
}

// setFocus sets focus to the specified panel.
func (a *App) setFocus(panel FocusPanel) {
	a.focusPanel = panel
	a.updateFocus()
}

// updateFocus updates the UI to reflect the current focus state.
func (a *App) updateFocus() {
	switch a.focusPanel {
	case FocusFeedList:
		a.tviewApp.SetFocus(a.feedList.list)
		a.feedList.list.SetBorderColor(tcell.ColorDarkCyan)
		a.articleList.list.SetBorderColor(tcell.ColorGray)
	case FocusArticleList:
		a.tviewApp.SetFocus(a.articleList.list)
		a.feedList.list.SetBorderColor(tcell.ColorGray)
		a.articleList.list.SetBorderColor(tcell.ColorDarkCyan)
	}
}

// setStatus updates the status bar text.
func (a *App) setStatus(message string) {
	a.statusBar.SetText(" " + message)
}

// Run starts the application and displays the splash screen.
func (a *App) Run() error {
	// Focus splash screen provider list
	a.tviewApp.SetFocus(a.splash.providerList)

	return a.tviewApp.Run()
}

// selectProvider handles provider selection from the splash screen.
func (a *App) selectProvider(providerName string) {
	a.selectedProvider = providerName
	provider := config.GetProvider(providerName)

	// Update feed list with provider's feeds
	a.feedList.SetFeeds(provider.FeedOrder)

	// Switch to main page
	a.pages.SwitchToPage("main")
	a.updateFocus()

	// Load feeds for the selected provider
	go a.loadProviderFeeds(provider)
}

// loadProviderFeeds fetches all feeds for the selected provider.
func (a *App) loadProviderFeeds(provider config.Provider) {
	a.setStatus("[yellow]Loading feeds...[white]")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results := a.feedFetcher.FetchAllFromMap(ctx, provider.Feeds)

	a.feedResultsMux.Lock()
	for _, result := range results {
		a.feedResults[result.FeedName] = result
	}
	a.feedResultsMux.Unlock()

	// Update UI on main thread
	a.tviewApp.QueueUpdateDraw(func() {
		// Select first feed
		if len(provider.FeedOrder) > 0 {
			a.selectFeed(provider.FeedOrder[0])
		}
		a.setStatus("Press [yellow]?[white] for help | [yellow]q[white] to quit")
	})
}

// RefreshFeeds reloads all feed data for the current provider.
func (a *App) RefreshFeeds() {
	if a.selectedProvider == "" {
		return
	}

	provider := config.GetProvider(a.selectedProvider)

	go func() {
		a.tviewApp.QueueUpdateDraw(func() {
			a.setStatus("[yellow]Refreshing feeds...[white]")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results := a.feedFetcher.FetchAllFromMap(ctx, provider.Feeds)

		a.feedResultsMux.Lock()
		for _, result := range results {
			a.feedResults[result.FeedName] = result
		}
		a.feedResultsMux.Unlock()

		a.tviewApp.QueueUpdateDraw(func() {
			// Refresh current feed view
			if a.currentFeed != "" {
				a.selectFeed(a.currentFeed)
			}
			a.setStatus("[green]Feeds refreshed[white] | Press [yellow]?[white] for help")
		})
	}()
}

// selectFeed updates the article list for the selected feed.
func (a *App) selectFeed(feedName string) {
	a.currentFeed = feedName

	a.feedResultsMux.RLock()
	result, exists := a.feedResults[feedName]
	a.feedResultsMux.RUnlock()

	if !exists || result.Error != nil {
		a.articleList.Clear()
		if result.Error != nil {
			a.setStatus("[red]Error loading " + feedName + ": " + result.Error.Error() + "[white]")
		}
		return
	}

	a.articleList.SetArticles(result.Feed.Channel.Items, feedName)
}

// openArticle fetches and displays the full article content.
func (a *App) openArticle(item rss.Item) {
	a.articleView.ShowLoading(item.Title)
	a.pages.SwitchToPage("article")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		fetchedArticle, err := a.articleFetcher.Fetch(ctx, item.Link)

		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				a.articleView.ShowError(item.Title, err)
			} else {
				a.articleView.ShowArticle(fetchedArticle)
			}
		})
	}()
}

// closeArticle returns to the main view.
func (a *App) closeArticle() {
	a.pages.SwitchToPage("main")
	a.updateFocus()
}
