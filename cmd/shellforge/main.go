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
	// configure logging
	logger, logPath, closeLog := configureApplicationLogging()
	defer closeLog()
	slog.SetDefault(logger)
	slog.Info("Shellforge started", "log_file", logPath)
	defer slog.Info("Shellforge stopped")

	// configure storage
	model, closeStorage := newApplicationModel()
	defer closeStorage()

	// run the TUI program
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
	// open the SQLite store
	store, err := storage.OpenDefault()
	// if store cannot be opened, run with an empty store
	// completions and settings will stay in memory only
	if err != nil {
		slog.Error("Shellforge storage could not start", logErrorKey, err)
		return app.New(), func() {}
	}
	// if store is opened successfully, run with it and ensure it is closed on exit
	return app.NewWithStorage(store), func() {
		if err := store.Close(); err != nil {
			slog.Error("Shellforge storage could not close", logErrorKey, err)
		}
	}
}

func runProgram(model app.State) error {
	program := tea.NewProgram(model)

	// this will block until the program exits, either normally or with an error
	finalModel, err := program.Run()

	fmt.Println("Thank you for using Shellforge!")

	finalState, ok := finalModel.(app.State)
	if !ok {
		slog.Error("Unexpected final model type; terminal not closed", "type", fmt.Sprintf("%T", finalModel))
	} else {
		finalState.CloseTerminal()
	}

	return err
}
