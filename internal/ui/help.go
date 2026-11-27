package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// keybinding describes a keyboard shortcut for the help screen.
type keybinding struct {
	key         string
	description string
}

// helpSection is a titled group of keybindings on the help screen.
type helpSection struct {
	title    string
	bindings []keybinding
}

// helpSections lists every shortcut, grouped by the screen it applies to.
// Ctrl+C, which quits from anywhere, is mentioned in the help footer.
var helpSections = []helpSection{
	{"Feeds and articles", []keybinding{
		{"j/k ↓/↑", "Move"},
		{"d/u", "Page down/up"},
		{"g/G", "Top/bottom"},
		{"h/l Tab", "Switch panel"},
		{"Enter", "Open article"},
		{"r", "Refresh feeds"},
		{"Esc", "Providers"},
		{"?", "Help"},
		{"q", "Quit"},
	}},
	{"Article", []keybinding{
		{"j/k", "Scroll"},
		{"d/u", "Half page"},
		{"Space/b", "Page down/up"},
		{"g/G", "Top/bottom"},
		{"o", "Open in browser"},
		{"?", "Help"},
		{"q/Esc", "Back"},
	}},
}

// renderHelpView draws the keybinding help box in the middle of a
// width×height screen.
func renderHelpView(width, height int) string {
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
	placed := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
	// Clip rather than overflow on terminals too small for the box.
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(placed)
}
