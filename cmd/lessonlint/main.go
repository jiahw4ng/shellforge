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

// validate checks cross-file lesson rules after the loader has parsed and
// structurally validated every individual YAML definition.
func validate(files fs.FS) error {
	available, err := lessons.LoadLessonsFromFile(files)
	if err != nil {
		return err
	}

	numbers := make(map[int]string, len(available))
	ids := make(map[string]string, len(available))
	for _, lesson := range available {
		if previous, exists := numbers[lesson.Number]; exists {
			return fmt.Errorf("lesson number %d is used by both %q and %q", lesson.Number, previous, lesson.ID)
		}
		if _, exists := ids[lesson.ID]; exists {
			return fmt.Errorf("lesson ID %q is used more than once", lesson.ID)
		}
		numbers[lesson.Number] = lesson.ID
		ids[lesson.ID] = lesson.ID
	}

	return nil
}
