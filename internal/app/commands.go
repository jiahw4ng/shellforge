package app

import (
	"context"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"

	tea "charm.land/bubbletea/v2"
)

// startTerminal creates a terminal session without blocking Bubble Tea's event loop.
// it will load the lesson, if any
func startTerminal(width, height int, lesson *lessons.Lesson) tea.Cmd {
	return func() tea.Msg {
		session, err := terminal.Start(context.Background(), width, height, lesson)
		return TerminalStartedMsg{Session: session, Err: err}
	}
}

// waitForTerminalExit converts the session's exit signal into a Bubble Tea message.
func waitForTerminalExit(exited <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-exited
		return TerminalExitedMsg{}
	}
}
