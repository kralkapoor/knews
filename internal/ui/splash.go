package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/config"
)

const minProviderListWidth = 30

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
	provider config.Provider
}

func (p providerItem) FilterValue() string { return p.provider.Name }

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
	base := lipgloss.NewStyle()
	if selected {
		base = selectedStyle
	}
	nameStyle := base.Bold(selected).PaddingLeft(2)
	descStyle := base.PaddingLeft(4)
	if !selected {
		descStyle = descStyle.Foreground(colorGray)
	}

	// Pad to full width
	availWidth := m.Width()
	name := nameStyle.Width(availWidth).Render(truncateString(pi.provider.Name, availWidth-2))
	desc := descStyle.Width(availWidth).Render(truncateString(pi.provider.Description, availWidth-4))

	fmt.Fprintf(w, "%s\n%s", name, desc)
}

type splashModel struct {
	providerList list.Model
	width        int
	height       int
}

func newSplashModel(providers []config.Provider) splashModel {
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = providerItem{provider: p}
	}

	// Wide enough for the longest description and its 4-column indent.
	width := minProviderListWidth
	for _, p := range providers {
		width = max(width, lipgloss.Width(p.Description)+4, lipgloss.Width(p.Name)+2)
	}

	l := newList(providerDelegate{}, width, len(providers)*2)
	l.SetItems(items)
	return splashModel{providerList: l}
}

func (s *splashModel) setSize(width, height int) {
	s.width = width
	s.height = height
}

func (s splashModel) Update(msg tea.Msg) (splashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return s, tea.Quit
		case "?":
			return s, func() tea.Msg { return showHelpMsg{} }
		case "enter":
			if item, ok := s.providerList.SelectedItem().(providerItem); ok {
				return s, func() tea.Msg { return providerSelectedMsg{provider: item.provider} }
			}
			return s, nil
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
	status := statusText(" ", "Enter", " Select  ", "?", " Help  ", "q", " Quit")

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
	contentHeight := max(s.height-lipgloss.Height(statusBox), 0)
	centeredContent := lipgloss.Place(s.width, contentHeight, lipgloss.Center, lipgloss.Center, content)

	view := lipgloss.JoinVertical(lipgloss.Left, centeredContent, statusBox)
	// Clip rather than overflow on terminals too small for the banner.
	return lipgloss.NewStyle().MaxWidth(s.width).MaxHeight(s.height).Render(view)
}
