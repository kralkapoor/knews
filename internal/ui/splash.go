package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/kralkapoor/knews/internal/config"
	"github.com/rivo/tview"
)

const knewsBanner = `
888  /   888b    | 888~~  Y88b         / ,d88~~\
888 /    |Y88b   | 888___  Y88b       /  8888   
888/\    | Y88b  | 888      Y88b  e  /   'Y88b  
888  \   |  Y88b | 888       Y88bd8b/     'Y88b,
888   \  |   Y88b| 888        Y88Y8Y        8888
888    \ |    Y888 888___      Y  Y      \__88P'
`

// SplashScreen represents the initial provider selection screen.
type SplashScreen struct {
	layout       *tview.Flex
	providerList *tview.List
	app          *App
}

// NewSplashScreen creates a new splash screen with provider selection.
func NewSplashScreen(app *App) *SplashScreen {
	splash := &SplashScreen{
		app: app,
	}

	splash.setupUI()
	return splash
}

// setupUI initializes the splash screen layout and components.
func (s *SplashScreen) setupUI() {
	// Create banner text view
	banner := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[darkcyan]" + knewsBanner + "[white]")

	// Create provider selection prompt
	prompt := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white]Select a news provider:[white]")

	// Create provider list
	s.providerList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add providers to the list
	providerNames := config.ProviderNames()
	for _, name := range providerNames {
		provider := config.GetProvider(name)
		s.providerList.AddItem(name, provider.Description, 0, nil)
	}

	// Handle provider selection
	s.providerList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		s.app.selectProvider(mainText)
	})

	// Create status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(" [yellow]Enter[white] Select  [yellow]q[white] Quit")
	statusBar.SetBorder(true).SetBorderColor(tcell.ColorGray)

	// Create centered list container
	listContainer := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(s.providerList, 30, 0, true).
		AddItem(nil, 0, 1, false)

	// Create main content area
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(banner, 8, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(prompt, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(listContainer, 5, 0, true).
		AddItem(nil, 0, 1, false)

	// Create outer layout with status bar
	s.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true).
		AddItem(statusBar, 3, 0, false)
}

// HandleInput processes keyboard input for the splash screen.
func (s *SplashScreen) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			s.app.tviewApp.Stop()
			return nil
		case 'j':
			// Move down in list
			current := s.providerList.GetCurrentItem()
			if current < s.providerList.GetItemCount()-1 {
				s.providerList.SetCurrentItem(current + 1)
			}
			return nil
		case 'k':
			// Move up in list
			current := s.providerList.GetCurrentItem()
			if current > 0 {
				s.providerList.SetCurrentItem(current - 1)
			}
			return nil
		}
	case tcell.KeyDown:
		current := s.providerList.GetCurrentItem()
		if current < s.providerList.GetItemCount()-1 {
			s.providerList.SetCurrentItem(current + 1)
		}
		return nil
	case tcell.KeyUp:
		current := s.providerList.GetCurrentItem()
		if current > 0 {
			s.providerList.SetCurrentItem(current - 1)
		}
		return nil
	case tcell.KeyEnter:
		index := s.providerList.GetCurrentItem()
		mainText, _ := s.providerList.GetItemText(index)
		s.app.selectProvider(mainText)
		return nil
	}
	return event
}
