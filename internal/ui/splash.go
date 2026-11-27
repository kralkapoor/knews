package ui

import (
	"strings"

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

// updateSplash handles a key press on the provider selection screen.
func (m *Model) updateSplash(key string) tea.Cmd {
	switch key {
	case "q":
		return tea.Quit
	case "enter":
		if len(m.providers) == 0 {
			return nil
		}
		return m.selectProvider(m.providers[m.providerList.cursor])
	default:
		// Every provider is on screen, so a page is the whole list.
		m.providerList.handleKey(key, len(m.providers), len(m.providers))
	}
	return nil
}

// providerListWidth is wide enough for the longest description and its
// 4-column indent.
func (m Model) providerListWidth() int {
	width := minProviderListWidth
	for _, p := range m.providers {
		width = max(width, lipgloss.Width(p.Description)+4, lipgloss.Width(p.Name)+2)
	}
	return width
}

// providerRow draws one provider as two lines: its name and its description.
func providerRow(p config.Provider, selected bool, width int) string {
	base := lipgloss.NewStyle()
	if selected {
		base = selectedStyle
	}
	nameStyle := base.Bold(selected).PaddingLeft(2)
	descStyle := base.PaddingLeft(4)
	if !selected {
		descStyle = descStyle.Foreground(colorGray)
	}

	// Pad each line to the full width so the selection highlight is a bar.
	name := nameStyle.Width(width).Render(truncateString(p.Name, width-2))
	desc := descStyle.Width(width).Render(truncateString(p.Description, width-4))
	return name + "\n" + desc
}

func (m Model) viewSplash() string {
	// Banner
	banner := bannerStyle.Render(knewsBanner)
	banner = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, banner)

	// Prompt
	prompt := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, "Select a news provider:")

	// Provider list (centered)
	listWidth := m.providerListWidth()
	rows := make([]string, len(m.providers))
	for i, p := range m.providers {
		rows[i] = providerRow(p, i == m.providerList.cursor, listWidth)
	}
	list := strings.Join(rows, "\n")
	if len(rows) == 0 {
		list = emptyListText
	}
	listView := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, list)

	// Status bar
	status := " " + keyHint("Enter", "Select") + "  " + keyHint("?", "Help") + "  " + keyHint("q", "Quit")
	statusBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGray).
		Width(max(m.width-2, 1)).
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
	contentHeight := max(m.height-lipgloss.Height(statusBox), 0)
	centeredContent := lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, content)

	view := lipgloss.JoinVertical(lipgloss.Left, centeredContent, statusBox)
	// Clip rather than overflow on terminals too small for the banner.
	return lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(view)
}
