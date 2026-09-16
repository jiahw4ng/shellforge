package lessons

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestLoadReadsEmbeddedLessons verifies the binary includes the first two
// version-controlled lesson definitions in numeric order.
func TestLoadReadsEmbeddedLessons(t *testing.T) {
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("Load() returned %d lessons, want 2", len(loaded))
	}
	if loaded[0].ID != "01-navigation" || loaded[1].ID != "02-files-and-directories" {
		t.Fatalf("Load() returned lesson IDs %q and %q", loaded[0].ID, loaded[1].ID)
	}
}

// TestLoadFromFSRejectsDuplicateNumbers prevents two files from occupying the
// same position in the lesson menu.
func TestLoadFromFSRejectsDuplicateNumbers(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml":  {Data: []byte("id: first\nnumber: 1\ntitle: First\ncontent: First content\n")},
		"02-second.yaml": {Data: []byte("id: second\nnumber: 1\ntitle: Second\ncontent: Second content\n")},
	}

	_, err := LoadFromFS(files)
	if err == nil || !strings.Contains(err.Error(), "lesson number 1") {
		t.Fatalf("LoadFromFS() error = %v, want duplicate-number error", err)
	}
}

// TestLoadFromFSRejectsIncompleteLesson keeps invalid content from starting a
// learner container later in the application flow.
func TestLoadFromFSRejectsIncompleteLesson(t *testing.T) {
	files := fstest.MapFS{
		"01-invalid.yaml": {Data: []byte("id: invalid\nnumber: 1\ntitle: Missing content\n")},
	}

	_, err := LoadFromFS(files)
	if err == nil || !strings.Contains(err.Error(), "missing content") {
		t.Fatalf("LoadFromFS() error = %v, want missing-content error", err)
	}
}
