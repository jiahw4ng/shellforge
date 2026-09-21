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
	logger, logFile, logPath, err := logging.Configure()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shellforge logging could not start:", err)
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	} else {
		defer func() { _ = logFile.Close() }()
	}
	slog.SetDefault(logger)
	slog.Info("Shellforge started", "log_file", logPath)
	defer slog.Info("Shellforge stopped")

	model := app.New()
	completionStore, completionErr := completion.OpenDefault()
	if completionErr != nil {
		slog.Error("Shellforge completion storage could not start", "error", completionErr)
	} else {
		defer func() {
			if err := completionStore.Close(); err != nil {
				slog.Error("Shellforge completion storage could not close", "error", err)
			}
		}()
		model = app.NewWithCompletionStore(completionStore)
	}

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	fmt.Println("Thank you for using Shellforge!")
	if model, ok := finalModel.(app.State); ok {
		model.Close()
	}
	if err != nil {
		slog.Error("Shellforge stopped with an error", "error", err)
		fmt.Fprintln(os.Stderr, "Shellforge could not start:", err)
		os.Exit(1)
	}
}
