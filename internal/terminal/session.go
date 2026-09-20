// Package terminal owns an interactive Bubbleterm session and its disposable
// Docker container.
package terminal

import (
	"context"
	"fmt"
	"log/slog"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"sync"

	bubbleterm "github.com/taigrr/bubbleterm"
)

// Start creates a disposable Docker sandbox and connects Bubbleterm to its shell.
func Start(ctx context.Context, width, height int, lesson *lessons.Lesson) (*TermSession, error) {
	slog.Info("starting lesson terminal", "width", width, "height", height)

	// create a disposable Docker container for the lesson
	sandbox, err := container.CreateAndStart(ctx)
	if err != nil {
		slog.Error("could not start sandbox/lesson container", "error", err)
		return nil, &TerminalStartError{Stage: "create lesson sandbox", Err: err}
	}
	// if a lesson is provided, run its setup commands in the container
	if lesson != nil {
		if err := sandbox.RunSetupLesson(ctx, lesson.Setup); err != nil {
			sandbox.Remove(context.Background())
			slog.Error("could not run lesson setup", "error", err)
			return nil, &TerminalStartError{Stage: "prepare lesson sandbox", Err: err}
		}
	}

	// start the Bubbleterm emulator with the container's shell command
	emulator, err := bubbleterm.NewWithCommand(
		ui.DimensionWithFallback(width, 80),
		ui.DimensionWithFallback(height, 24),
		sandbox.ShellCommand(),
	)
	if err != nil {
		slog.Error("could not start terminal emulator", "error", err)
		sandbox.Remove(context.Background())
		return nil, &TerminalStartError{Stage: "start terminal emulator", Err: err}
	}

	// create a channel that closes when the shell process exits
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

// NewInvalidStartResultError reports an impossible terminal-start result from
// a command: it supplied neither a session nor an error.
func NewInvalidStartResultError() error {
	return fmt.Errorf("terminal startup returned no session and no error")
}
