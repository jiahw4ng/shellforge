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

func TestStorePersistsCursorStyleAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), databaseFileName)
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if styleID, found, err := store.CursorStyle(t.Context()); err != nil || found {
		t.Fatalf("initial CursorStyle() = %q, %v, %v; want not found", styleID, found, err)
	}
	if err := store.MarkCompleted(t.Context(), "00-introduction"); err != nil {
		t.Fatalf("MarkCompleted() error = %v", err)
	}
	for _, styleID := range []string{"bar-steady", "underline-blinking"} {
		if err := store.SaveCursorStyle(t.Context(), styleID); err != nil {
			t.Fatalf("SaveCursorStyle(%q) error = %v", styleID, err)
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

	styleID, found, err := reopened.CursorStyle(t.Context())
	if err != nil || !found || styleID != "underline-blinking" {
		t.Fatalf("CursorStyle() after reopen = %q, %v, %v; want underline-blinking", styleID, found, err)
	}
	completed, err := reopened.CompletedLessonIDs(t.Context())
	if err != nil {
		t.Fatalf("CompletedLessonIDs() error = %v", err)
	}
	if want := []string{"00-introduction"}; !reflect.DeepEqual(completed, want) {
		t.Fatalf("completed lessons = %v, want %v", completed, want)
	}
}
