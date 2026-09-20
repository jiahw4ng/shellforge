// Package screens contains Shellforge's presentation-only TUI renderers.
package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// MainMenu renders the application greeting and main navigation choices.
func MainMenu(items []string, selection int) string {
	lines := []string{ui.TitleStyle.Render(MainMenuTitle), ""}
	for index, item := range items {
		if index == len(items)-1 {
			lines = append(lines, "", selectableItem(index, selection, item))
		} else {
			lines = append(lines, selectableItem(index, selection, item))
		}
	}

	lines = append(lines, "", ui.MutedStyle.Render(MenuNavigationPrompt))
	return strings.Join(lines, "\n")
}
