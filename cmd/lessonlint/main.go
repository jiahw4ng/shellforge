// Command lessonlint validates Shellforge's embedded lesson definitions.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"shellforge/internal/lessons"

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
		if err := lesson.Validate(); err != nil {
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
