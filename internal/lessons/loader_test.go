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

// TestLoadFromFSRejectsIncompleteLesson keeps invalid content from starting a
// learner container later in the application flow.
func TestLoadFromFSRejectsIncompleteLesson(t *testing.T) {
	files := fstest.MapFS{
		"01-invalid.yaml": {Data: []byte("id: invalid\nnumber: 1\ntitle: Missing pages\n")},
	}

	_, err := LoadLessonsFromFile(files)
	if err == nil || !strings.Contains(err.Error(), "at least one page") {
		t.Fatalf("LoadLessonsFromFile() error = %v, want missing-pages error", err)
	}
}
