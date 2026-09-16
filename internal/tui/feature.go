package tui

import "strings"

// displayFeatureView generates the placeholder view shown after a menu selection.
func displayFeatureView() string {
	return strings.Join([]string{
		titleStyle.Render("feature coming soon!"),
		"",
		selected.Render("> Back"),
		"",
		muted.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
