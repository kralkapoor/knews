package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/config"
)

const knewsBanner = `
888  /   888b    | 888~~  Y88b         / ,d88~~\
888 /    |Y88b   | 888___  Y88b       /  8888
888/\    | Y88b  | 888      Y88b  e  /   'Y88b
888  \   |  Y88b | 888       Y88bd8b/     'Y88b,
888   \  |   Y88b| 888        Y88Y8Y        8888
888    \ |    Y888 888___      Y  Y      \__88P'
`

// providerItem implements list.Item for the provider selection list.
type providerItem struct {
	name        string
	description string
}

func (p providerItem) Title() string       { return p.name }
func (p providerItem) Description() string { return p.description }
func (p providerItem) FilterValue() string { return p.name }

// providerDelegate renders items in the provider list.
type providerDelegate struct{}

func (d providerDelegate) Height() int                             { return 2 }
func (d providerDelegate) Spacing() int                            { return 0 }
func (d providerDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d providerDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	pi, ok := item.(providerItem)
	if !ok {
		return
	}

	selected := index == m.Index()

	nameStyle := lipgloss.NewStyle().PaddingLeft(2)
	descStyle := lipgloss.NewStyle().PaddingLeft(4).Foreground(colorGray)

	if selected {
		nameStyle = nameStyle.Foreground(lipgloss.Color("15")).Background(colorDarkCyan).Bold(true)
		descStyle = descStyle.Foreground(lipgloss.Color("15")).Background(colorDarkCyan)
	}

	// Pad to full width
	availWidth := m.Width()
	name := nameStyle.Width(availWidth).Render(pi.name)
	desc := descStyle.Width(availWidth).Render(pi.description)

	fmt.Fprintf(w, "%s\n%s", name, desc)
}

type splashModel struct {
	providerList list.Model
	width        int
	height       int
}

func newSplashModel() splashModel {
	providers := config.ProviderNames()
	items := make([]list.Item, len(providers))
	for i, name := range providers {
		p := config.GetProvider(name)
		items[i] = providerItem{name: name, description: p.Description}
	}

	l := list.New(items, providerDelegate{}, 30, len(providers)*2+1)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return splashModel{
		providerList: l,
	}
}

func (s splashModel) Update(msg tea.Msg) (splashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return s, tea.Quit
		case "enter":
			item, ok := s.providerList.SelectedItem().(providerItem)
			if ok {
				return s, func() tea.Msg { return providerSelectedMsg{name: item.name} }
			}
		}
	}

	// Forward all other messages (j/k/up/down) to the list
	var cmd tea.Cmd
	s.providerList, cmd = s.providerList.Update(msg)
	return s, cmd
}

func (s splashModel) View() string {
	// Banner
	banner := bannerStyle.Render(knewsBanner)
	banner = lipgloss.PlaceHorizontal(s.width, lipgloss.Center, banner)

	// Prompt
	prompt := lipgloss.PlaceHorizontal(s.width, lipgloss.Center, "Select a news provider:")

	// Provider list (centered)
	listView := lipgloss.PlaceHorizontal(s.width, lipgloss.Center, s.providerList.View())

	// Status bar
	status := statusText(" ", "Enter", " Select  ", "q", " Quit")

	statusBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGray).
		Width(s.width - 2).
		Render(status)

	// Compose vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		banner,
		"",
		prompt,
		"",
		listView,
	)

	// Center content vertically, then add status bar at bottom
	contentHeight := s.height - lipgloss.Height(statusBox) - 1
	centeredContent := lipgloss.Place(s.width, contentHeight, lipgloss.Center, lipgloss.Center, content)

	return lipgloss.JoinVertical(lipgloss.Left, centeredContent, statusBox)
}
