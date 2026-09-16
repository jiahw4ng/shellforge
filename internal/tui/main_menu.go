package tui

import (
	"shellforge/internal/ui"
	"strings"
)

// New creates the initial main-menu model
func New() State {
	return State{}
}

// displayMenuViewWithSelectArrow generates the view for the main menu screen, highlighting the currently selected item
func displayMenuViewWithSelectArrow(selection int) string {
	lines := []string{ui.TitleStyle.Render("Welcome to Shellforge!"), ""}
	for index, item := range menuItems {
		prefix := "  "
		if index == selection {
			prefix = "> "
			item = ui.SelectedStyle.Render(item)
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "", ui.MutedStyle.Render("Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."))
	return strings.Join(lines, "\n")
}
