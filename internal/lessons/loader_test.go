package lessons

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestLoadReadsEmbeddedLessons verifies the binary includes the first four
// version-controlled lesson definitions in numeric order.
func TestLoadReadsEmbeddedLessons(t *testing.T) {
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 4 {
		t.Fatalf("Load() returned %d lessons, want 4", len(loaded))
	}
	if loaded[0].ID != "00-debugging" || loaded[1].ID != "01-introduction" || loaded[2].ID != "02-navigation" || loaded[3].ID != "03-files-and-directories" {
		t.Fatalf("Load() returned lesson IDs %q, %q, %q, and %q", loaded[0].ID, loaded[1].ID, loaded[2].ID, loaded[3].ID)
	}
	foundBashPage := false
	for _, page := range loaded[1].Pages {
		if strings.Contains(string(page.Content), "Bourne Again SHell") {
			foundBashPage = true
			break
		}
	}
	if !foundBashPage {
		t.Fatal("introduction lesson does not contain Bash content")
	}
}

// TestLoadLessonsFromFileRejectsMissingPageMarkdown prevents a valid YAML
// definition from referring to an omitted embedded page file.
func TestLoadLessonsFromFileRejectsMissingPageMarkdown(t *testing.T) {
	files := fstest.MapFS{
		"01-invalid.yaml": {Data: []byte("id: invalid\nnumber: 1\ntitle: Missing page\npages:\n  - title: Page\n    file: missing.md\n")},
	}

	_, err := LoadLessonsFromFile(files)
	if err == nil || !strings.Contains(err.Error(), "read missing.md") {
		t.Fatalf("LoadLessonsFromFile() error = %v, want missing-page-file error", err)
	}
}

// TestLoadLessonsFromFileDoesNotValidateLessonMetadata keeps authored-content
// policy in cmd/lessonlint rather than the application startup path.
func TestLoadLessonsFromFileDoesNotValidateLessonMetadata(t *testing.T) {
	files := fstest.MapFS{
		"01-invalid.yaml": {Data: []byte("id: invalid\nnumber: 1\ntitle: Missing pages\n")},
	}

	loaded, err := LoadLessonsFromFile(files)
	if err != nil {
		t.Fatalf("LoadLessonsFromFile() error = %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "invalid" {
		t.Fatalf("LoadLessonsFromFile() = %#v, want unvalidated lesson", loaded)
	}
}
