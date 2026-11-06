// Package logging configures Shellforge's file-based diagnostic logger.
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
)

const (
	fileName                  = "shellforge.log"
	localStateDirectory       = ".local"
	stateDirectory            = "state"
	applicationStateDirectory = "shellforge"
)

// activePath is the log file most recently opened by configureAt.
var activePath atomic.Pointer[string]

// Path returns the log file Configure opened, or an empty string when
// diagnostics are not being written to a file, so callers never point the
// learner at a log that does not exist.
func Path() string {
	if path := activePath.Load(); path != nil {
		return *path
	}
	return ""
}

// Configure creates Shellforge's private log file and returns a logger that
// writes only to that file, never to the Bubble Tea terminal UI.
func Configure() (*slog.Logger, io.Closer, string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, "", err
	}

	logDirectory := filepath.Join(homeDirectory, localStateDirectory, stateDirectory, applicationStateDirectory)
	return configureAt(filepath.Join(logDirectory, fileName))
}

// configureAt creates a logger at one specific path. Keeping this separate
// lets tests use a temporary directory instead of the user's real log file.
func configureAt(path string) (*slog.Logger, io.Closer, string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, nil, "", err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, "", err
	}

	handler := slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug})
	activePath.Store(&path)
	return slog.New(handler), file, path, nil
}
