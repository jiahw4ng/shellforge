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

func TestValidateAcceptsCommandHistoryAssertion(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: command_history_contains\n    contains: cd project\n")},
		"first.md":      {Data: []byte("Content")},
	}

	if err := validate(files); err != nil {
		t.Fatalf("validate() error = %v, want command-history assertion accepted", err)
	}
}

func TestValidateRejectsCommandHistoryAssertionWithoutCommandText(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: command_history_contains\n")},
		"first.md":      {Data: []byte("Content")},
	}

	err := validate(files)
	if err == nil || !strings.Contains(err.Error(), "requires command text") {
		t.Fatalf("validate() error = %v, want missing-command-text error", err)
	}
}

func TestValidateAcceptsFileModeAssertion(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: file_mode\n    path: scripts/run.sh\n    mode: \"755\"\n")},
		"first.md":      {Data: []byte("Content")},
	}

	if err := validate(files); err != nil {
		t.Fatalf("validate() error = %v, want file-mode assertion accepted", err)
	}
}

func TestValidateRejectsFileModeAssertionWithInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "75", "0755", "758"} {
		t.Run(mode, func(t *testing.T) {
			files := fstest.MapFS{
				"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: file_mode\n    path: scripts/run.sh\n    mode: \"" + mode + "\"\n")},
				"first.md":      {Data: []byte("Content")},
			}

			err := validate(files)
			if err == nil || !strings.Contains(err.Error(), "requires a three-digit octal mode") {
				t.Fatalf("validate() error = %v, want three-digit-octal-mode error", err)
			}
		})
	}
}

func TestValidateAcceptsCurrentWorkingDirectoryAssertion(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: cwd\n    path: /home/student/workspace/project/docs\n")},
		"first.md":      {Data: []byte("Content")},
	}

	if err := validate(files); err != nil {
		t.Fatalf("validate() error = %v, want current-working-directory assertion accepted", err)
	}
}

func TestValidateRejectsCurrentWorkingDirectoryAssertionWithNonNormalizedPath(t *testing.T) {
	for _, assertionPath := range []string{"project/docs", "/home/student/workspace/project/../docs"} {
		t.Run(assertionPath, func(t *testing.T) {
			files := fstest.MapFS{
				"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: cwd\n    path: " + assertionPath + "\n")},
				"first.md":      {Data: []byte("Content")},
			}

			err := validate(files)
			if err == nil || !strings.Contains(err.Error(), "requires a normalized absolute path") {
				t.Fatalf("validate() error = %v, want normalized-absolute-path error", err)
			}
		})
	}
}

func TestValidateAcceptsEnvironmentVariableAssertion(t *testing.T) {
	files := fstest.MapFS{
		"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: environment_variable_exists\n    name: EDITOR\n")},
		"first.md":      {Data: []byte("Content")},
	}

	if err := validate(files); err != nil {
		t.Fatalf("validate() error = %v, want environment-variable assertion accepted", err)
	}
}

func TestValidateRejectsEnvironmentVariableAssertionWithInvalidName(t *testing.T) {
	for _, name := range []string{"", "EDITOR-NAME", "1EDITOR"} {
		t.Run(name, func(t *testing.T) {
			files := fstest.MapFS{
				"01-first.yaml": {Data: []byte("id: first\nnumber: 1\ntitle: First\npages:\n  - title: Page\n    file: first.md\nassertions:\n  - type: environment_variable_exists\n    name: " + name + "\n")},
				"first.md":      {Data: []byte("Content")},
			}

			err := validate(files)
			if err == nil || !strings.Contains(err.Error(), "requires a valid environment variable name") {
				t.Fatalf("validate() error = %v, want valid-environment-variable-name error", err)
			}
		})
	}
}
