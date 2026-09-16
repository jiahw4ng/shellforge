package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// TerminalStart renders the terminal's loading or recoverable error state.
func TerminalStart(startError error) string {
	if startError == nil {
		return ui.MutedStyle.Render("Starting sandboxed shell...")
	}

	return strings.Join([]string{
		ui.TitleStyle.Render("Unable to start sandboxed Bash"),
		startError.Error(),
		"",
		ui.MutedStyle.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
