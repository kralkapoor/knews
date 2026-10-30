package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	width  int
	height int
}

func (h *helpModel) setSize(width, height int) {
	h.width = width
	h.height = height
}

func (h helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return h, func() tea.Msg { return closeHelpMsg{} }
	}
	return h, nil
}

func (h helpModel) View() string {
	columns := make([]string, len(helpSections))
	for i, section := range helpSections {
		var b strings.Builder
		b.WriteString(keyStyle.Render(section.title))
		for _, kb := range section.bindings {
			fmt.Fprintf(&b, "\n%-8s %s", kb.key, kb.description)
		}
		columns[i] = lipgloss.NewStyle().PaddingRight(4).Render(b.String())
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Keyboard Shortcuts"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, columns...),
		"",
		metadataStyle.Render("Ctrl+C quits from anywhere."),
		metadataStyle.Render("Press any key to close."),
	)
	content = lipgloss.NewStyle().PaddingLeft(1).Render(content)

	box := titledBox(content, "Help", lipgloss.Width(content)+2, lipgloss.Height(content)+2, colorYellow)
	placed := lipgloss.Place(h.width, h.height, lipgloss.Center, lipgloss.Center, box)
	// Clip rather than overflow on terminals too small for the box.
	return lipgloss.NewStyle().MaxWidth(h.width).MaxHeight(h.height).Render(placed)
}
