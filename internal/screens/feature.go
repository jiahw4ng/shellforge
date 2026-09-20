package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// Feature renders the placeholder view shown for unfinished menu options.
func Feature() string {
	return strings.Join([]string{
		ui.TitleStyle.Render("Feature coming soon!"),
		"",
		ui.SelectedStyle.Render("> Back"),
		"",
		ui.MutedStyle.Render(MenuNavigationPrompt),
	}, "\n")
}
