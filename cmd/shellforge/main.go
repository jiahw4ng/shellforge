package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"shellforge/internal/logging"
	"shellforge/internal/tui"

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

	program := tea.NewProgram(tui.New())
	finalModel, err := program.Run()
	if model, ok := finalModel.(tui.State); ok {
		model.Close()
	}
	if err != nil {
		slog.Error("Shellforge stopped with an error", "error", err)
		fmt.Fprintln(os.Stderr, "Shellforge could not start:", err)
		os.Exit(1)
	}
}
