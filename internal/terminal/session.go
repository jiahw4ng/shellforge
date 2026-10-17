// Package terminal owns an interactive virtual-terminal session and its disposable
// Docker container.
package terminal

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"

	"github.com/charmbracelet/x/xpty"
)

const (
	createSandboxStage  = "create lesson sandbox"
	prepareSandboxStage = "prepare lesson sandbox"
	startEmulatorStage  = "start terminal emulator"
	logErrorKey         = "error"
)

// Start creates a disposable Docker sandbox and connects a terminal engine to its shell.
func Start(ctx context.Context, width, height int, lesson *lessons.Lesson) (*TermSession, error) {
	slog.Info("starting lesson terminal", "width", width, "height", height)

	sandbox, err := prepareSandbox(ctx, lesson)
	if err != nil {
		return nil, err
	}

	terminal, command, err := startTerminal(width, height, sandbox)
	if err != nil {
		return nil, err
	}

	processContext, cancelProcess := context.WithCancel(context.Background())
	return &TermSession{
		terminal:      terminal,
		sandbox:       sandbox,
		cancelProcess: cancelProcess,
		process:       command.Process,
		Exited:        observeProcessExit(processContext, command),
	}, nil
}

func prepareSandbox(ctx context.Context, lesson *lessons.Lesson) (*container.Container, error) {
	sandbox, err := container.CreateAndStart(ctx)
	if err != nil {
		slog.Error("could not start sandbox/lesson container", logErrorKey, err)
		return nil, &TerminalStartError{Stage: createSandboxStage, Err: err}
	}
	if lesson == nil {
		return sandbox, nil
	}
	if err := sandbox.RunSetupLesson(ctx, lesson.Setup); err != nil {
		sandbox.Remove(context.Background())
		slog.Error("could not run lesson setup", logErrorKey, err)
		return nil, &TerminalStartError{Stage: prepareSandboxStage, Err: err}
	}
	return sandbox, nil
}

func startTerminal(width, height int, sandbox *container.Container) (*terminalEngine, *exec.Cmd, error) {
	terminalWidth := ui.DimensionWithFallback(width, 80)
	terminalHeight := ui.DimensionWithFallback(height, 24)
	pseudoTerminal, err := xpty.NewPty(terminalWidth, terminalHeight)
	if err != nil {
		slog.Error("could not start terminal emulator", logErrorKey, err)
		sandbox.Remove(context.Background())
		return nil, nil, &TerminalStartError{Stage: startEmulatorStage, Err: err}
	}

	command := sandbox.ShellCommand()
	preparePTYCommand(command)
	if err := pseudoTerminal.Start(command); err != nil {
		_ = pseudoTerminal.Close()
		slog.Error("could not start terminal emulator", logErrorKey, err)
		sandbox.Remove(context.Background())
		return nil, nil, &TerminalStartError{Stage: startEmulatorStage, Err: err}
	}
	return newTerminalEngine(terminalWidth, terminalHeight, pseudoTerminal), command, nil
}

func observeProcessExit(ctx context.Context, command *exec.Cmd) <-chan struct{} {
	exited := make(chan struct{})
	go func() {
		if err := xpty.WaitProcess(ctx, command); err != nil && ctx.Err() == nil {
			slog.Debug("lesson Bash session exited with an error", logErrorKey, err)
		}
		slog.Info("lesson Bash session exited")
		close(exited)
	}()
	return exited
}

// NewInvalidStartResultError reports an impossible terminal-start result from
// a command: it supplied neither a session nor an error.
func NewInvalidStartResultError() error {
	return fmt.Errorf("terminal startup returned no session and no error")
}
