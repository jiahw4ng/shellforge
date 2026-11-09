package screens

import (
	"shellforge/internal/container"
	"shellforge/internal/ui"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// TerminalPane renders a consistently labelled sandbox above either a live
// terminal frame or one of the terminal's loading and failure states.
func TerminalPane(content string, width, height int) string {
	width = ui.DimensionWithFallback(width, 80)
	height = ui.DimensionWithFallback(height, 24)
	header := ui.ControlStyle.Render(sandboxTitle)
	body := lipgloss.NewStyle().
		Width(width).
		Height(TerminalContentHeight(height)).
		Render(content)
	return lipgloss.NewStyle().Width(width).Height(height).Render(header + "\n" + body)
}

// TerminalContentHeight returns the rows available to the terminal after the
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
func TerminalStart(startError error, steps []container.Step, spinnerState string) string {
	if startError != nil {
		return terminalError(startError.Error())
	}

	lines := StartupLines(steps)
	if len(lines) == 0 {
		lines = []string{startingSandbox}
	}
	if spinnerState != "" {
		lines[len(lines)-1] += " " + spinnerState
	}
	return ui.MutedStyle.Render(strings.Join(lines, "\n"))
}

// StartupLines describes the startup steps reached so far, one line per step,
// in the order they happened. The wording of the container step depends on
// whether the image had to be downloaded first.
func StartupLines(steps []container.Step) []string {
	pulled := slices.Contains(steps, container.StepPullingImage)
	lines := make([]string, 0, len(steps))
	for _, step := range steps {
		switch step {
		case container.StepCheckingDocker:
			lines = append(lines, checkingDockerLine)
		case container.StepCheckingImage:
			lines = append(lines, checkingImageLine)
		case container.StepPullingImage:
			lines = append(lines, pullingImageLine)
		case container.StepStartingContainer:
			if pulled {
				lines = append(lines, startingAfterPullLine)
			} else {
				lines = append(lines, startingImageFoundLine)
			}
		case container.StepPreparingLesson:
			lines = append(lines, preparingLessonLine)
		case container.StepConnectingShell:
			lines = append(lines, connectingShellLine)
		}
	}
	return lines
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

// TerminalProblem is the learner-facing explanation of a terminal that could
// not start. Every field except Summary is optional.
type TerminalProblem struct {
	// Summary says what went wrong; Remedy says what to do about it.
	Summary string
	Remedy  string
	// Hint says how to leave the error screen.
	Hint string
	// Detail quotes the underlying tool's own message.
	Detail string
	// LogPath is where full diagnostics are written; empty when there is no log file.
	LogPath string
}

// View renders a recognised startup failure as a short summary and remedy,
// keeping the underlying message and the log location secondary.
func (p TerminalProblem) View() string {
	lines := []string{
		ui.ErrorStyle.Render(terminalErrorTitle),
		"",
		ui.TextStyle.Render(p.Summary),
	}
	if p.Remedy != "" {
		lines = append(lines, ui.TextStyle.Render(p.Remedy))
	}
	if p.Hint != "" {
		lines = append(lines, "", ui.TextStyle.Render(p.Hint))
	}
	if p.Detail != "" || p.LogPath != "" {
		lines = append(lines, "")
	}
	if p.Detail != "" {
		lines = append(lines, ui.MutedStyle.Render(p.Detail))
	}
	if p.LogPath != "" {
		lines = append(lines, ui.MutedStyle.Render(logPathPrefix+p.LogPath))
	}
	return strings.Join(lines, "\n")
}

// terminalError renders a recoverable terminal error consistently in both the
// standalone sandbox and the lesson terminal pane.
func terminalError(details string) string {
	return strings.Join([]string{
		ui.ErrorStyle.Render(terminalErrorTitle),
		"",
		ui.MutedStyle.Render(details),
	}, "\n")
}
