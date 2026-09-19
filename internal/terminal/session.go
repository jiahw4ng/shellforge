// Package terminal owns an interactive Bubbleterm session and its disposable
// Docker container.
package terminal

import (
	"context"
	"log/slog"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	bubbleterm "github.com/taigrr/bubbleterm"
)

// Start creates a disposable Docker sandbox and connects Bubbleterm to its shell.
func Start(ctx context.Context, width, height int, lesson *lessons.Lesson) (*TermSession, error) {
	slog.Info("starting lesson terminal", "width", width, "height", height)

	// create a disposable Docker container for the lesson
	sandbox, err := container.CreateAndStart(ctx)
	if err != nil {
		slog.Error("could not start sandbox/lesson container", "error", err)
		return nil, err
	}
	// if a lesson is provided, run its setup commands in the container
	if lesson != nil {
		if err := sandbox.RunSetupLesson(ctx, lesson.Setup); err != nil {
			sandbox.Remove(context.Background())
			slog.Error("could not run lesson setup", "error", err)
			return nil, err
		}
	}

	emulator, err := bubbleterm.NewWithCommand(
		dimension(width, 80),
		dimension(height, 24),
		sandbox.ShellCommand(),
	)
	if err != nil {
		slog.Error("could not start terminal emulator", "error", err)
		sandbox.Remove(context.Background())
		return nil, err
	}

	exited := make(chan struct{})
	var notifyExit sync.Once
	notify := func(string) {
		notifyExit.Do(func() {
			slog.Info("lesson Bash session exited")
			close(exited)
		})
	}
	emulator.GetEmulator().SetOnExit(notify)
	if emulator.GetEmulator().IsProcessExited() {
		notify("")
	}

	return &TermSession{emulator: emulator, sandbox: sandbox, exited: exited}, nil
}

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

func dimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
