package tui

import "strings"

// featureView generates the placeholder view shown after a menu selection.
func featureView() string {
	return strings.Join([]string{
		titleStyle.Render("feature coming soon!"),
		"",
		selected.Render("> Back"),
		"",
		muted.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
