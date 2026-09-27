package lessons

import (
	"fmt"
	"io/fs"
	"path"
)

// Validate checks one lesson's metadata and page material. The lessonlint
// command invokes it during development and CI, before lessons are embedded in
// a binary. Lesson 0 is reserved for the introduction.
func (lesson Lesson) Validate() error {
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
		if err := assertion.Validate(); err != nil {
			return fmt.Errorf("lesson %q assertion %d: %w", lesson.ID, index+1, err)
		}
	}

	return nil
}
