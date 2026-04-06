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

func newHelpModel() helpModel {
	return helpModel{}
}

func (h helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return h, func() tea.Msg { return closeHelpMsg{} }
	}
	return h, nil
}

func (h helpModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Navigation:"))
	b.WriteString("\n")
	for _, kb := range NavigationBindings() {
		b.WriteString(fmt.Sprintf("  %-10s %s\n", kb.Key, kb.Description))
	}

	b.WriteString("\n")
	b.WriteString(keyStyle.Render("Actions:"))
	b.WriteString("\n")
	for _, kb := range ActionBindings() {
		b.WriteString(fmt.Sprintf("  %-10s %s\n", kb.Key, kb.Description))
	}

	b.WriteString("\n")
	b.WriteString(metadataStyle.Render("Press any key to close"))

	content := b.String()

	boxWidth := 50
	boxHeight := lipgloss.Height(content) + 2 // +2 for borders
	box := titledBox(content, "Help", boxWidth, boxHeight, colorYellow)

	return lipgloss.Place(h.width, h.height, lipgloss.Center, lipgloss.Center, box)
}
