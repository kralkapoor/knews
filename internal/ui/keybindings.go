package ui

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
