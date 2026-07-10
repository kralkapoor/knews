package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors for the dark cyan / gray scheme.
var (
	colorDarkCyan = lipgloss.Color("#008b8b")
	colorGray     = lipgloss.Color("240")
	colorYellow   = lipgloss.Color("3")
	colorRed      = lipgloss.Color("1")
	colorGreen    = lipgloss.Color("2")
)

// Text styles used across screens.
var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	keyStyle      = lipgloss.NewStyle().Foreground(colorYellow)
	metadataStyle = lipgloss.NewStyle().Foreground(colorGray)
	errorStyle    = lipgloss.NewStyle().Foreground(colorRed)
	successStyle  = lipgloss.NewStyle().Foreground(colorGreen)
	bannerStyle   = lipgloss.NewStyle().Foreground(colorDarkCyan)
)

// titledBox renders content inside a bordered box with an optional title in
// the top border, using Unicode box-drawing characters.
//
//	╭─ Title ──────────╮
//	│ content           │
//	╰───────────────────╯
func titledBox(content, title string, width, height int, borderColor lipgloss.Color) string {
	bc := lipgloss.NewStyle().Foreground(borderColor)

	// Inner width excludes the two border columns ("│" on each side).
	inner := width - 2
	if inner < 1 {
		inner = 1
	}

	// Build top border
	var top string
	if title != "" {
		titleStr := " " + title + " "
		titleRendered := bc.Render(titleStr)
		dashCount := inner - lipgloss.Width(titleStr)
		if dashCount < 0 {
			dashCount = 0
		}
		top = bc.Render("╭─") + titleRendered + bc.Render(strings.Repeat("─", dashCount)+"╮")
	} else {
		top = bc.Render("╭" + strings.Repeat("─", inner) + "╮")
	}

	// Build bottom border
	bottom := bc.Render("╰" + strings.Repeat("─", inner) + "╯")

	// Pad/truncate content lines to fill the box
	contentLines := strings.Split(content, "\n")

	// Inner height is total height minus top and bottom borders
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	// Pad or truncate to innerHeight lines
	for len(contentLines) < innerHeight {
		contentLines = append(contentLines, "")
	}
	if len(contentLines) > innerHeight {
		contentLines = contentLines[:innerHeight]
	}

	var lines []string
	lines = append(lines, top)
	for _, line := range contentLines {
		// Pad line to inner width
		lineWidth := lipgloss.Width(line)
		padding := inner - lineWidth
		if padding < 0 {
			padding = 0
		}
		lines = append(lines, bc.Render("│")+line+strings.Repeat(" ", padding)+bc.Render("│"))
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

// statusText builds a status bar string with highlighted keys.
// parts alternate: normal text, key text, normal text, key text, ...
func statusText(parts ...string) string {
	var result string
	for i, part := range parts {
		if i%2 == 1 {
			result += keyStyle.Render(part)
		} else {
			result += part
		}
	}
	return result
}
