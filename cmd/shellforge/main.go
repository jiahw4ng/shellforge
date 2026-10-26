package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"shellforge/internal/app"
	"shellforge/internal/logging"
	"shellforge/internal/storage"

	tea "charm.land/bubbletea/v2"
)

const logErrorKey = "error"

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

	model, closeStorage := newApplicationModel()
	defer closeStorage()

	if err := runProgram(model); err != nil {
		slog.Error("Shellforge stopped with an error", logErrorKey, err)
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
	store, err := storage.OpenDefault()
	if err != nil {
		slog.Error("Shellforge storage could not start", logErrorKey, err)
		return app.New(), func() {}
	}
	return app.NewWithStorage(store), func() {
		if err := store.Close(); err != nil {
			slog.Error("Shellforge storage could not close", logErrorKey, err)
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
