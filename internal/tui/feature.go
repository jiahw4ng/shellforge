package tui

import (
	"shellforge/internal/ui"
	"strings"
)

// displayFeatureView generates the placeholder view shown after a menu selection.
func displayFeatureView() string {
	return strings.Join([]string{
		ui.TitleStyle.Render("feature coming soon!"),
		"",
		ui.SelectedStyle.Render("> Back"),
		"",
		ui.MutedStyle.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
