package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors for the dark cyan / gray scheme.
var (
	colorDarkCyan = lipgloss.Color("#008b8b")
	colorGray     = lipgloss.Color("240")
	colorWhite    = lipgloss.Color("15")
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
	selectedStyle = lipgloss.NewStyle().Background(colorDarkCyan).Foreground(colorWhite)
)

// titledBox renders content inside a width×height bordered box with an
// optional title in the top border. Content and title are clipped to fit.
//
//	╭─ Title ──────────╮
//	│ content          │
//	╰──────────────────╯
func titledBox(content, title string, width, height int, borderColor lipgloss.Color) string {
	bc := lipgloss.NewStyle().Foreground(borderColor)

	// Inner width excludes the two border columns ("│" on each side).
	inner := max(width-2, 1)

	// The top border is "╭─", the title padded with a space each side, then
	// dashes up to "╮". Drop the title entirely if there's no room for it.
	title = truncateString(title, inner-3)
	var top string
	if title != "" {
		titleStr := " " + title + " "
		dashCount := max(inner-1-lipgloss.Width(titleStr), 0)
		top = bc.Render("╭─" + titleStr + strings.Repeat("─", dashCount) + "╮")
	} else {
		top = bc.Render("╭" + strings.Repeat("─", inner) + "╮")
	}
	bottom := bc.Render("╰" + strings.Repeat("─", inner) + "╯")

	// Pad or clip content to the inner height (total minus both borders).
	innerHeight := max(height-2, 0)
	contentLines := strings.Split(content, "\n")
	for len(contentLines) < innerHeight {
		contentLines = append(contentLines, "")
	}
	contentLines = contentLines[:innerHeight]

	clip := lipgloss.NewStyle().MaxWidth(inner)
	lines := make([]string, 0, innerHeight+2)
	lines = append(lines, top)
	for _, line := range contentLines {
		line = clip.Render(line)
		padding := max(inner-lipgloss.Width(line), 0)
		lines = append(lines, bc.Render("│")+line+strings.Repeat(" ", padding)+bc.Render("│"))
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

// keyHint renders a key in the highlight colour followed by what it does,
// e.g. keyHint("q", "Quit") is "q Quit" with the "q" highlighted.
func keyHint(key, label string) string {
	return keyStyle.Render(key) + " " + label
}

// truncateString trims s and shortens it to at most maxLength runes, ending
// with "..." when there's room. It counts runes rather than bytes so
// multi-byte characters (e.g. the macrons in te reo Māori) are never split.
func truncateString(s string, maxLength int) string {
	s = strings.TrimSpace(s)
	if maxLength <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	if maxLength < 3 {
		return string(runes[:maxLength])
	}
	return string(runes[:maxLength-3]) + "..."
}
