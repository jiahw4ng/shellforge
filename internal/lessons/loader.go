// Package lessons loads and validates Shellforge's embedded lesson definitions.
package lessons

import (
	"fmt"
	"io/fs"
	"sort"

	lessondata "shellforge/lessons"

	"gopkg.in/yaml.v3"
)

// Load parses the lesson YAML files embedded in the Shellforge binary.
func Load() ([]Lesson, error) {
	return LoadLessonsFromFile(lessondata.Files)
}

// LoadLessonsFromFile parses lesson YAML files from any filesystem. It is exported so
// callers and tests can load an alternate lesson set without touching disk.
func LoadLessonsFromFile(files fs.FS) ([]Lesson, error) {
	paths, err := fs.Glob(files, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("find lesson files: %w", err)
	}
	sort.Strings(paths)

	lessons := make([]Lesson, 0, len(paths))
	for _, path := range paths {
		lesson, err := loadLessonFromFile(files, path)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

// loadLessonFromFile decodes and validates one lesson YAML document.
func loadLessonFromFile(files fs.FS, path string) (Lesson, error) {
	contents, err := fs.ReadFile(files, path)
	if err != nil {
		return Lesson{}, fmt.Errorf("read %s: %w", path, err)
	}

	var lesson Lesson
	if err := yaml.Unmarshal(contents, &lesson); err != nil {
		return Lesson{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := lesson.validate(); err != nil {
		return Lesson{}, fmt.Errorf("validate %s: %w", path, err)
	}
	return lesson, nil
}
