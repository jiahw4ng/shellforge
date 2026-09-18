// Package lessons loads Shellforge's embedded lesson definitions.
package lessons

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	lessondata "shellforge/lessons"

	"gopkg.in/yaml.v3"
)

// Load parses and hydrates the lesson files embedded in the Shellforge binary.
func Load() ([]Lesson, error) {
	return LoadLessonsFromFile(lessondata.Files)
}

// Parse reads lesson YAML files from any filesystem without validating their
// authored material or reading their referenced Markdown. lessonlint uses it to
// validate lesson definitions during development and CI.
func Parse(files fs.FS) ([]Lesson, error) {
	paths, err := fs.Glob(files, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("find lesson files: %w", err)
	}
	sort.Strings(paths)

	lessons := make([]Lesson, 0, len(paths))
	for _, path := range paths {
		lesson, err := parseLessonFromFile(files, path)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

// LoadLessonsFromFile parses lesson YAML files and reads their referenced
// Markdown content. It intentionally does not validate authored lesson
// material; cmd/lessonlint owns those checks.
func LoadLessonsFromFile(files fs.FS) ([]Lesson, error) {
	lessons, err := Parse(files)
	if err != nil {
		return nil, err
	}

	for index := range lessons {
		if err := loadPageMarkdown(files, &lessons[index]); err != nil {
			return nil, fmt.Errorf("load pages for lesson %q: %w", lessons[index].ID, err)
		}
	}

	return lessons, nil
}

// parseLessonFromFile decodes one YAML document without inspecting its fields.
func parseLessonFromFile(files fs.FS, path string) (Lesson, error) {
	contents, err := fs.ReadFile(files, path)
	if err != nil {
		return Lesson{}, fmt.Errorf("read %s: %w", path, err)
	}

	var lesson Lesson
	if err := yaml.Unmarshal(contents, &lesson); err != nil {
		return Lesson{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return lesson, nil
}

// loadPageMarkdown reads the Markdown referenced by each page. Semantic lesson
// validation belongs to cmd/lessonlint; this function only hydrates content.
func loadPageMarkdown(files fs.FS, lesson *Lesson) error {
	for index := range lesson.Pages {
		page := &lesson.Pages[index]
		contents, err := fs.ReadFile(files, page.File)
		if err != nil {
			return fmt.Errorf("page %d: read %s: %w", index+1, page.File, err)
		}
		page.Content = Markdown(strings.TrimSpace(string(contents)))
	}
	return nil
}
