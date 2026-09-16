package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// Feature renders the placeholder view shown for unfinished menu options.
func Feature() string {
	return strings.Join([]string{
		ui.TitleStyle.Render("feature coming soon!"),
		"",
		ui.SelectedStyle.Render("> Back"),
		"",
		ui.MutedStyle.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
