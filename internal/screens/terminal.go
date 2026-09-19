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

	return terminalError(startError.Error())
}

// TerminalFailure keeps the terminal's final output visible after an
// unexpected exit, so Docker and Bash errors remain readable to the learner.
func TerminalFailure(output string, err error) string {
	details := strings.TrimSpace(output)
	if details == "" && err != nil {
		details = err.Error()
	}
	if details == "" {
		details = "The terminal closed unexpectedly."
	}
	return terminalError(details)
}

// terminalError renders a recoverable terminal error consistently in both the
// standalone sandbox and the lesson terminal pane.
func terminalError(details string) string {
	return strings.Join([]string{
		ui.ErrorStyle.Render("Oops! There was an error."),
		"",
		ui.MutedStyle.Render(details),
	}, "\n")
}
