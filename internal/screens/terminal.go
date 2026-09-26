package screens

import (
	"shellforge/internal/ui"
	"strings"

	"charm.land/lipgloss/v2"
)

const terminalHeaderHeight = 1

// TerminalPane renders a consistently labelled sandbox above either a live
// Bubbleterm frame or one of the terminal's loading and failure states.
func TerminalPane(content string, width, height int) string {
	width = ui.DimensionWithFallback(width, 80)
	height = ui.DimensionWithFallback(height, 24)
	header := ui.MutedStyle.Render("Sandboxed Bash")
	body := lipgloss.NewStyle().
		Width(width).
		Height(TerminalContentHeight(height)).
		Render(content)
	return lipgloss.NewStyle().Width(width).Height(height).Render(header + "\n" + body)
}

// TerminalContentHeight returns the rows available to Bubbleterm after the
// terminal pane's header has been accounted for.
func TerminalContentHeight(height int) int {
	height = ui.DimensionWithFallback(height, 24) - terminalHeaderHeight
	if height < 1 {
		return 1
	}
	return height
}

// TerminalStart renders the terminal's loading or recoverable error state, with
// a spinner if the terminal is still starting.
func TerminalStart(startError error, spinnerState string) string {
	if startError == nil {
		if spinnerState != "" {
			return ui.MutedStyle.Render(spinnerState + " Starting sandboxed shell...")
		}
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
		details = "An unknown error caused the terminal to close unexpectedly."
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
