package tui

import (
	"context"
	"shellforge/internal/container"
	"strings"

	tea "charm.land/bubbletea/v2"
	bubbleterm "github.com/taigrr/bubbleterm"
)

type terminalStartedMsg struct {
	terminal        *bubbleterm.Model
	lessonContainer *container.LessonContainer
	exited          <-chan struct{}
	err             error
}

type terminalExitedMsg struct{}

// startTerminal creates a disposable Docker lesson container, then gives its
// interactive shell command to Bubbleterm. Bubbleterm owns the PTY and VT
// emulation; the outer model owns the Docker container lifetime.
func startTerminal(width, height int) tea.Cmd {
	return func() tea.Msg {
		lessonContainer, err := container.CreateAndStart(context.Background())
		if err != nil {
			return terminalStartedMsg{err: err}
		}

		terminal, err := bubbleterm.NewWithCommand(
			terminalDimension(width, 80),
			terminalDimension(height, 24),
			lessonContainer.ShellCommand(),
		)
		if err != nil {
			lessonContainer.Remove(context.Background())
			return terminalStartedMsg{err: err}
		}

		exited := make(chan struct{}, 1)
		terminal.GetEmulator().SetOnExit(func(string) {
			exited <- struct{}{}
		})
		if terminal.GetEmulator().IsProcessExited() {
			exited <- struct{}{}
		}

		return terminalStartedMsg{
			terminal:        terminal,
			lessonContainer: lessonContainer,
			exited:          exited,
		}
	}
}

func waitForTerminalExit(exited <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-exited
		return terminalExitedMsg{}
	}
}

func (m *Model) handleTerminalStarted(message terminalStartedMsg) {
	if message.err != nil {
		m.terminalStartErr = message.err
		return
	}

	m.terminal = message.terminal
	m.lessonContainer = message.lessonContainer
	m.terminalExit = message.exited
	m.terminalStartErr = nil
}

func (m Model) updateTerminal(message tea.Msg) (tea.Model, tea.Cmd) {
	updated, command := m.terminal.Update(message)
	m.terminal = updated.(*bubbleterm.Model)
	return m, command
}

func (m *Model) closeTerminal() {
	if m.terminal != nil {
		_ = m.terminal.Close()
	}
	if m.lessonContainer != nil {
		m.lessonContainer.Remove(context.Background())
	}
	m.terminal = nil
	m.lessonContainer = nil
	m.terminalExit = nil
}

// Close releases an active terminal and removes its disposable lesson container.
// It is safe to call after the terminal has already exited.
func (m *Model) Close() {
	m.closeTerminal()
}

func terminalDimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func terminalStartView(startError error) string {
	if startError == nil {
		return muted.Render("Starting sandboxed shell...")
	}

	return strings.Join([]string{
		titleStyle.Render("Unable to start sandboxed Bash"),
		startError.Error(),
		"",
		muted.Render("Press Enter to return. Ctrl+C exits."),
	}, "\n")
}
