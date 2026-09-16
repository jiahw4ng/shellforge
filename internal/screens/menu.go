// Package screens contains Shellforge's presentation-only TUI renderers.
package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// MainMenu renders the application greeting and main navigation choices.
func MainMenu(items []string, selection int) string {
	lines := []string{ui.TitleStyle.Render("Welcome to Shellforge!"), ""}
	for index, item := range items {
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
