package app

import (
	"context"
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
	"time"

	tea "charm.land/bubbletea/v2"
)

// startTerminal creates a terminal session without blocking Bubble Tea's event loop.
// it will load the lesson, if any
func startTerminal(width, height int, lesson *lessons.Lesson) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		session, err := terminal.Start(ctx, width, height, lesson)
		return TerminalStartedMsg{Session: session, Err: err}
	}
}

// checkAssertions evaluates lesson outcomes without blocking the TUI or PTY.
func checkAssertions(session *terminal.TermSession, assertions []assertion.Assertion) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return AssertionsCheckedMsg{Results: session.EvaluateAssertions(ctx, assertions)}
	}
}

// waitForTerminalExit converts the session's exit signal into a Bubble Tea message.
func waitForTerminalExit(exited <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-exited
		return TerminalExitedMsg{}
	}
}
