package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateRejectsDuplicateLessonNumbers(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml":  {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    content: First content\n")},
		"02-second.yaml": {Data: []byte("id: second\nnumber: 1\ntitle: Second\npages:\n  - title: Page\n    content: Second content\n")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "lesson number 1") {
		t.Fatalf("validate() error = %v, want duplicate-number error", err)
	}
}

func TestValidateRejectsDuplicateLessonIDs(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml":  {Data: []byte("id: same\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    content: First content\n")},
		"02-second.yaml": {Data: []byte("id: same\nnumber: 2\ntitle: Second\npages:\n  - title: Page\n    content: Second content\n")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "lesson ID") {
		t.Fatalf("validate() error = %v, want duplicate-ID error", err)
	}
}
