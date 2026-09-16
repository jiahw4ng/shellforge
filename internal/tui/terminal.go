package tui

import (
	"context"
	"log/slog"
	"shellforge/internal/container"
	"shellforge/internal/ui"
	"strings"

	tea "charm.land/bubbletea/v2"
	bubbleterm "github.com/taigrr/bubbleterm"
)

type terminalStartedMsg struct {
	terminal        *bubbleterm.Model
	lessonContainer *container.Container
	exited          <-chan struct{}
	err             error
}

type terminalExitedMsg struct{}

// startTerminal creates a disposable Docker lesson container, then gives its
// interactive shell command to Bubbleterm. Bubbleterm owns the PTY and VT
// emulation; the outer model owns the Docker container lifetime.
func startTerminal(width, height int) tea.Cmd {
	return func() tea.Msg {
		slog.Info("starting lesson terminal", "width", width, "height", height)
		lessonContainer, err := container.CreateAndStart(context.Background())
		if err != nil {
			slog.Error("could not start lesson terminal", "error", err)
			return terminalStartedMsg{err: err}
		}

		terminal, err := bubbleterm.NewWithCommand(
			terminalDimension(width, 80),
			terminalDimension(height, 24),
			lessonContainer.ShellCommand(),
		)
		if err != nil {
			slog.Error("could not start terminal emulator", "error", err)
			lessonContainer.Remove(context.Background())
			return terminalStartedMsg{err: err}
		}

		exited := make(chan struct{}, 1)

		// this means, once the terminal exits, we will send a message to the "exited" channel
		terminal.GetEmulator().SetOnExit(func(string) {
			slog.Info("lesson Bash session exited")
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

// this function is used to wait for the terminal to exit, and then send a message to the main model
func waitForTerminalExit(exited <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		// this means to wait for a message to be received on the "exited" channel, which means the terminal has exited
		// then we will return terminalExitedMsg{}
		<-exited
		return terminalExitedMsg{}
	}
}

// handleTerminalStarted records either the ready terminal resources or the
// startup error returned by the asynchronous launch command.
func (m *State) handleTerminalStarted(message terminalStartedMsg) {
	if message.err != nil {
		slog.Error("lesson terminal failed to start", "error", message.err)
		m.TermErr = message.err
		return
	}

	m.Terminal = message.terminal
	m.LessonContainer = message.lessonContainer
	m.TermExit = message.exited
	m.TermErr = nil
}

// updateTerminal passes a Bubble Tea event to Bubbleterm and saves its updated
// model plus any follow-up command it needs to run.
func (m State) updateTerminal(message tea.Msg) (tea.Model, tea.Cmd) {
	updated, command := m.Terminal.Update(message)
	m.Terminal = updated.(*bubbleterm.Model)
	return m, command
}

// closeTerminal stops the emulator and force-removes the Docker container that
// belongs to this lesson, then clears the terminal-related UI state.
func (m *State) closeTerminal() {
	slog.Info("closing lesson terminal")
	if m.Terminal != nil {
		_ = m.Terminal.Close()
	}
	if m.LessonContainer != nil {
		m.LessonContainer.Remove(context.Background())
	}
	m.Terminal = nil
	m.LessonContainer = nil
	m.TermExit = nil
}

// Close releases an active terminal and removes its disposable lesson container.
// It is safe to call after the terminal has already exited.
func (m *State) Close() {
	m.closeTerminal()
}

// terminalDimension substitutes a safe default before Bubbleterm has received
// the outer terminal's first size event.
func terminalDimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

// terminalStartView renders either the brief loading state or a recoverable
// error screen while no terminal emulator is available.
func terminalStartView(startError error) string {
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
