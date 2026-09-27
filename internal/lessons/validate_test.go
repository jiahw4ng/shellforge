package lessons

import (
	"strings"
	"testing"
)

func TestLessonValidateRequiresSuccessMessage(t *testing.T) {
	lesson := Lesson{
		ID:     "lesson-id",
		Number: 1,
		Title:  "Lesson title",
		Pages:  []Page{{Title: "Page title", File: "page.md"}},
	}

	err := lesson.Validate()
	if err == nil || !strings.Contains(err.Error(), "missing success message") {
		t.Fatalf("Validate() error = %v, want missing-success-message error", err)
	}
}
