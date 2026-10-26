package terminal

import (
	"context"
	"log/slog"
	"shellforge/internal/assertion"

	tea "charm.land/bubbletea/v2"
)

// Init waits for the terminal engine's next output notification.
func (s *TermSession) Init() tea.Cmd {
	return s.terminal.waitForChange()
}

// Update forwards one Bubble Tea event to the terminal engine.
func (s *TermSession) Update(message tea.Msg) tea.Cmd {
	switch message := message.(type) {
	case tea.KeyPressMsg:
		return func() tea.Msg {
			s.terminal.sendKey(message)
			return nil
		}
	case terminalChangedMsg:
		if message.terminal == s.terminal {
			return s.terminal.waitForChange()
		}
	}
	return nil
}

// Resize changes the emulator's virtual terminal dimensions.
func (s *TermSession) Resize(width, height int) tea.Cmd {
	return func() tea.Msg {
		if err := s.terminal.resize(width, height); err != nil {
			slog.Debug("could not resize terminal", logErrorKey, err)
		}
		return nil
	}
}

// SendInput queues text for the terminal. It is useful for automated integration tests.
func (s *TermSession) SendInput(input string) tea.Cmd {
	return func() tea.Msg {
		s.terminal.sendText(input)
		return nil
	}
}

// EvaluateAssertions checks the active lesson's observable outcomes inside
// this session's disposable container.
func (s *TermSession) EvaluateAssertions(ctx context.Context, assertions []assertion.Assertion) []assertion.Result {
	return assertion.Evaluate(ctx, s.sandbox, assertions)
}

// View returns the terminal engine's current contents.
func (s *TermSession) View() string {
	return s.terminal.view()
}

// Cursor returns the visible terminal-local cursor as a blinking block.
func (s *TermSession) Cursor() *tea.Cursor {
	x, y, visible := s.terminal.cursor()
	if !visible {
		return nil
	}
	return &tea.Cursor{
		Position: tea.Position{X: x, Y: y},
		Shape:    tea.CursorBlock,
		Blink:    true,
	}
}

// Close stops the emulator and removes its disposable Docker container.
func (s *TermSession) Close() {
	s.closeOnce.Do(func() {
		slog.Info("closing lesson terminal")
		s.cancelProcess()
		_ = s.process.Kill()
		_ = s.terminal.close()
		<-s.Exited
		s.sandbox.Remove(context.Background())
	})
}
