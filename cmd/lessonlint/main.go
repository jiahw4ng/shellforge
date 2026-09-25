// Command lessonlint validates Shellforge's embedded lesson definitions.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"strings"

	lessondata "shellforge/lessons"
)

func main() {
	if err := validate(lessondata.Files); err != nil {
		fmt.Fprintln(os.Stderr, "lesson validation failed:", err)
		os.Exit(1)
	}
}

// validate checks every lesson-material rule after the loader has parsed and
// hydrated the embedded YAML and Markdown files.
func validate(files fs.FS) error {
	available, err := lessons.Parse(files)
	if err != nil {
		return err
	}

	numbers := make(map[int]string, len(available))
	ids := make(map[string]string, len(available))
	for _, lesson := range available {
		if err := validateLesson(lesson); err != nil {
			return err
		}
		if previous, exists := numbers[lesson.Number]; exists {
			return fmt.Errorf("lesson number %d is used by both %q and %q", lesson.Number, previous, lesson.ID)
		}
		if _, exists := ids[lesson.ID]; exists {
			return fmt.Errorf("lesson ID %q is used more than once", lesson.ID)
		}
		numbers[lesson.Number] = lesson.ID
		ids[lesson.ID] = lesson.ID

		pageFiles := make(map[string]struct{}, len(lesson.Pages))
		for _, page := range lesson.Pages {
			if _, exists := pageFiles[page.File]; exists {
				return fmt.Errorf("lesson %q references page file %q more than once", lesson.ID, page.File)
			}
			pageFiles[page.File] = struct{}{}
		}
	}

	hydrated, err := lessons.LoadLessonsFromFile(files)
	if err != nil {
		return err
	}
	for _, lesson := range hydrated {
		for index, page := range lesson.Pages {
			if page.Content == "" {
				return fmt.Errorf("lesson %q page %d: empty Markdown file %q", lesson.ID, index+1, page.File)
			}
		}
	}

	return nil
}

// validateLesson checks one lesson's metadata and page material. It lives in
// this command rather than the runtime loader so invalid authored content is
// rejected during development and CI, before it is embedded in a binary.
// Lesson 0 is reserved for the introduction.
func validateLesson(lesson lessons.Lesson) error {
	switch {
	case lesson.ID == "":
		return fmt.Errorf("lesson is missing an ID")
	case lesson.Number < 0:
		return fmt.Errorf("lesson %q: number must not be negative", lesson.ID)
	case lesson.Title == "":
		return fmt.Errorf("lesson %q: missing title", lesson.ID)
	case len(lesson.Pages) == 0:
		return fmt.Errorf("lesson %q: must contain at least one page", lesson.ID)
	}

	for index, page := range lesson.Pages {
		if page.Title == "" {
			return fmt.Errorf("lesson %q page %d: missing title", lesson.ID, index+1)
		}
		if page.File == "" {
			return fmt.Errorf("lesson %q page %d: missing file", lesson.ID, index+1)
		}
		if !fs.ValidPath(page.File) || path.Ext(page.File) != ".md" {
			return fmt.Errorf("lesson %q page %d: invalid Markdown file %q", lesson.ID, index+1, page.File)
		}
	}

	for index, assertion := range lesson.Assertions {
		if err := validateAssertion(assertion); err != nil {
			return fmt.Errorf("lesson %q assertion %d: %w", lesson.ID, index+1, err)
		}
	}

	return nil
}

// validateAssertion rejects unsupported assertion types and incomplete
// assertion data before a lesson is embedded in the application binary.
func validateAssertion(check assertion.Assertion) error {
	switch check.Type {
	case assertion.AssertionTypeDirectoryExists, assertion.AssertionTypeFileExists:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf("%s requires a path", check.Type)
		}
	case assertion.AssertionTypeFileContent, assertion.AssertionTypeFileContains:
		if strings.TrimSpace(check.Path) == "" {
			return fmt.Errorf("%s requires a path", check.Type)
		}
		if check.Contains == "" {
			return fmt.Errorf("%s requires content to find", check.Type)
		}
	case assertion.AssertionTypeCommandHistoryContains:
		if check.Contains == "" {
			return fmt.Errorf("%s requires command text to find", check.Type)
		}
	default:
		return fmt.Errorf("unsupported assertion type %q", check.Type)
	}

	return nil
}
