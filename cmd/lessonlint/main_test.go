package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateRejectsDuplicateLessonNumbers(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml":  {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\n")},
		"02-second.yaml": {Data: []byte("id: second\nnumber: 1\ntitle: Second\npages:\n  - title: Page\n    file: second.md\n")},
		"first.md":       {Data: []byte("First content")},
		"second.md":      {Data: []byte("Second content")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "lesson number 1") {
		t.Fatalf("validate() error = %v, want duplicate-number error", err)
	}
}

func TestValidateRejectsDuplicateLessonIDs(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml":  {Data: []byte("id: same\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\n")},
		"02-second.yaml": {Data: []byte("id: same\nnumber: 2\ntitle: Second\npages:\n  - title: Page\n    file: second.md\n")},
		"first.md":       {Data: []byte("First content")},
		"second.md":      {Data: []byte("Second content")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "lesson ID") {
		t.Fatalf("validate() error = %v, want duplicate-ID error", err)
	}
}

func TestValidateRejectsLessonWithoutPages(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\n")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "must contain at least one page") {
		t.Fatalf("validate() error = %v, want missing-pages error", err)
	}
}

func TestValidateRejectsInvalidPageFile(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.txt\n")},
		"first.txt":     {Data: []byte("Content")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "invalid Markdown file") {
		t.Fatalf("validate() error = %v, want invalid-page-file error", err)
	}
}

func TestValidateRejectsEmptyMarkdownPage(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\n")},
		"first.md":      {Data: []byte(" \n")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "empty Markdown file") {
		t.Fatalf("validate() error = %v, want empty-Markdown error", err)
	}
}
