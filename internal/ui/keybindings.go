package ui

// Keybinding defines a keyboard shortcut with its description.
type Keybinding struct {
	Key         string
	Description string
}

// NavigationBindings returns the navigation-related keybindings.
func NavigationBindings() []Keybinding {
	return []Keybinding{
		{Key: "j / ↓", Description: "Move down"},
		{Key: "k / ↑", Description: "Move up"},
		{Key: "h / ←", Description: "Focus feed list"},
		{Key: "l / →", Description: "Focus article list"},
		{Key: "Tab", Description: "Switch panels"},
		{Key: "Enter", Description: "Select item / Open article"},
		{Key: "g", Description: "Go to top (article view)"},
		{Key: "G", Description: "Go to bottom (article view)"},
		{Key: "PgUp", Description: "Page up"},
		{Key: "PgDn", Description: "Page down"},
	}
}

// ActionBindings returns the action-related keybindings.
func ActionBindings() []Keybinding {
	return []Keybinding{
		{Key: "r", Description: "Refresh feeds"},
		{Key: "o", Description: "Open article in browser"},
		{Key: "q", Description: "Quit / Back"},
		{Key: "?", Description: "Show help"},
		{Key: "Esc", Description: "Close article view"},
	}
}

// AllBindings returns all keybindings grouped by category.
func AllBindings() map[string][]Keybinding {
	return map[string][]Keybinding{
		"Navigation": NavigationBindings(),
		"Actions":    ActionBindings(),
	}
}
