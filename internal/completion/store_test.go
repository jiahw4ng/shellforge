package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestStorePersistsCompletedLessonsAcrossReopen(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "state")
	path := filepath.Join(directory, databaseFileName)

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	initial, err := store.CompletedLessonIDs(t.Context())
	if err != nil {
		t.Fatalf("CompletedLessonIDs() error = %v", err)
	}
	if len(initial) != 0 {
		t.Fatalf("initial completed lessons = %v, want none", initial)
	}

	for _, lessonID := range []string{"01-navigation", "00-introduction", "01-navigation"} {
		if err := store.MarkCompleted(t.Context(), lessonID); err != nil {
			t.Fatalf("MarkCompleted(%q) error = %v", lessonID, err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	t.Cleanup(func() { _ = reopened.Close() })

	completed, err := reopened.CompletedLessonIDs(t.Context())
	if err != nil {
		t.Fatalf("CompletedLessonIDs() after reopen error = %v", err)
	}
	want := []string{"00-introduction", "01-navigation"}
	if !reflect.DeepEqual(completed, want) {
		t.Fatalf("completed lessons = %v, want %v", completed, want)
	}

	if runtime.GOOS != "windows" {
		assertPermissions(t, directory, 0o700)
		assertPermissions(t, path, 0o600)
	}
}

func TestStoreResetsLessonCompletions(t *testing.T) {
	path := filepath.Join(t.TempDir(), databaseFileName)
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	for _, lessonID := range []string{"00-introduction", "01-navigation"} {
		if err := store.MarkCompleted(t.Context(), lessonID); err != nil {
			t.Fatalf("MarkCompleted(%q) error = %v", lessonID, err)
		}
	}
	if err := store.ResetLessonCompletions(t.Context()); err != nil {
		t.Fatalf("ResetLessonCompletions() error = %v", err)
	}
	completed, err := store.CompletedLessonIDs(t.Context())
	if err != nil {
		t.Fatalf("CompletedLessonIDs() error = %v", err)
	}
	if len(completed) != 0 {
		t.Fatalf("completed lessons after reset = %v, want none", completed)
	}
}

func assertPermissions(t *testing.T, path string, wantMode uint32) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if got := uint32(info.Mode().Perm()); got != wantMode {
		t.Fatalf("permissions for %q = %#o, want %#o", path, got, wantMode)
	}
}
