package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigureAtWritesPrivateLogFile confirms diagnostics are persisted to the
// requested file rather than written to Shellforge's terminal UI.
func TestConfigureAtWritesPrivateLogFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", fileName)
	logger, closer, returnedPath, err := configureAt(path)
	if err != nil {
		t.Fatalf("configureAt() error = %v", err)
	}
	t.Cleanup(func() { _ = closer.Close() })

	logger.Log(t.Context(), slog.LevelInfo, "test event", "key", "value")
	if err := closer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if returnedPath != path || !strings.Contains(string(contents), "msg=\"test event\"") {
		t.Fatalf("log contents = %q, path = %q", contents, returnedPath)
	}
}
