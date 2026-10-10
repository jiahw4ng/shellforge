// Package terminal owns an interactive Bubbleterm session and its disposable
// Docker container.
package terminal

import (
	"context"
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

	sandbox, err := prepareSandbox(ctx, lesson)
	if err != nil {
		return nil, err
	}

	emulator, err := startEmulator(width, height, sandbox)
	if err != nil {
		return nil, err
	}

	return &TermSession{emulator: emulator, sandbox: sandbox, Exited: observeProcessExit(emulator)}, nil
}

func prepareSandbox(ctx context.Context, lesson *lessons.Lesson) (*container.Container, error) {
	sandbox, err := container.CreateAndStart(ctx)
	if err != nil {
		slog.Error("could not start sandbox/lesson container", "error", err)
		return nil, &StartError{Stage: "create lesson sandbox", Err: err}
	}
	if lesson == nil {
		return sandbox, nil
	}
	if err := sandbox.RunSetupLesson(ctx, lesson.Setup); err != nil {
		sandbox.Remove(context.Background())
		slog.Error("could not run lesson setup", "error", err)
		return nil, &StartError{Stage: "prepare lesson sandbox", Err: err}
	}
	return sandbox, nil
}

func startEmulator(width, height int, sandbox *container.Container) (*bubbleterm.Model, error) {
	emulator, err := bubbleterm.NewWithCommand(
		ui.DimensionWithFallback(width, 80),
		ui.DimensionWithFallback(height, 24),
		sandbox.ShellCommand(),
	)
	if err != nil {
		slog.Error("could not start terminal emulator", "error", err)
		sandbox.Remove(context.Background())
		return nil, &StartError{Stage: "start terminal emulator", Err: err}
	}
	return emulator, nil
}

func observeProcessExit(emulator *bubbleterm.Model) <-chan struct{} {
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
	return exited
}
