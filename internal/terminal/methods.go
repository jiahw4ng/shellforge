package terminal

import (
	"context"
	"log/slog"
	"shellforge/internal/assertion"
	"strings"

	tea "charm.land/bubbletea/v2"
	bubbleterm "github.com/taigrr/bubbleterm"
)

// Init starts Bubbleterm's next asynchronous read command.
func (s *TermSession) Init() tea.Cmd {
	return s.emulator.Init()
}

// Update forwards one Bubble Tea event to Bubbleterm and retains its new model.
func (s *TermSession) Update(message tea.Msg) tea.Cmd {
	updated, command := s.emulator.Update(message)
	s.emulator = updated.(*bubbleterm.Model)
	return command
}

// Resize changes the emulator's virtual terminal dimensions.
func (s *TermSession) Resize(width, height int) tea.Cmd {
	return s.emulator.Resize(width, height)
}

// SendInput queues text for the terminal. It is useful for automated integration tests.
func (s *TermSession) SendInput(input string) tea.Cmd {
	return s.emulator.SendInput(input)
}

// EvaluateAssertions checks the active lesson's observable outcomes inside
// this session's disposable container.
func (s *TermSession) EvaluateAssertions(ctx context.Context, assertions []assertion.Assertion) []assertion.Result {
	return assertion.Evaluate(ctx, s.sandbox, assertions)
}

// View returns Bubbleterm's current terminal contents.
func (s *TermSession) View() string {
	frame := s.emulator.GetEmulator().GetScreen()
	return strings.Join(frame.Rows, "\n")
}

// Exited is closed when the shell process ends.
func (s *TermSession) Exited() <-chan struct{} {
	return s.exited
}

// Close stops the emulator and removes its disposable Docker container.
func (s *TermSession) Close() {
	slog.Info("closing lesson terminal")
	if s.emulator != nil {
		_ = s.emulator.Close()
	}
	if s.sandbox != nil {
		s.sandbox.Remove(context.Background())
	}
}
