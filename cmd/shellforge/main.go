package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"shellforge/internal/app"
	"shellforge/internal/completion"
	"shellforge/internal/logging"

	tea "charm.land/bubbletea/v2"
)

// main starts the TUI and ensures an active lesson container is cleaned up
// before Shellforge exits, whether normally or because Bubble Tea returns an error.
func main() {
	os.Exit(run())
}

func run() int {
	logger, logPath, closeLog := configureApplicationLogging()
	defer closeLog()
	slog.SetDefault(logger)
	slog.Info("Shellforge started", "log_file", logPath)
	defer slog.Info("Shellforge stopped")

	model, closeCompletionStore := newApplicationModel()
	defer closeCompletionStore()

	if err := runProgram(model); err != nil {
		slog.Error("Shellforge stopped with an error", "error", err)
		fmt.Fprintln(os.Stderr, "Shellforge could not start:", err)
		return 1
	}
	return 0
}

func configureApplicationLogging() (*slog.Logger, string, func()) {
	logger, logFile, logPath, err := logging.Configure()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shellforge logging could not start:", err)
		return slog.New(slog.NewTextHandler(io.Discard, nil)), "", func() {}
	}
	return logger, logPath, func() { _ = logFile.Close() }
}

func newApplicationModel() (app.State, func()) {
	model := app.New()
	completionStore, completionErr := completion.OpenDefault()
	if completionErr != nil {
		slog.Error("Shellforge completion storage could not start", "error", completionErr)
		return model, func() {}
	}
	return app.NewWithCompletionStore(completionStore), func() {
		if err := completionStore.Close(); err != nil {
			slog.Error("Shellforge completion storage could not close", "error", err)
		}
	}
}

func runProgram(model app.State) error {
	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	fmt.Println("Thank you for using Shellforge!")
	if finalState, ok := finalModel.(app.State); ok {
		finalState.CloseTerminal()
	}
	return err
}
