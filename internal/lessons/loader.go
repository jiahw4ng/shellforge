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
	return LoadFromFS(lessondata.Files)
}

// LoadFromFS parses lesson YAML files from any filesystem. It is exported so
// callers and tests can load an alternate lesson set without touching disk.
func LoadFromFS(files fs.FS) ([]Lesson, error) {
	paths, err := fs.Glob(files, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("find lesson files: %w", err)
	}
	sort.Strings(paths)

	lessons := make([]Lesson, 0, len(paths))
	numbers := make(map[int]string, len(paths))
	ids := make(map[string]string, len(paths))
	for _, path := range paths {
		lesson, err := loadFile(files, path)
		if err != nil {
			return nil, err
		}
		if previous, exists := numbers[lesson.Number]; exists {
			return nil, fmt.Errorf("%s: lesson number %d is already used by %s", path, lesson.Number, previous)
		}
		if previous, exists := ids[lesson.ID]; exists {
			return nil, fmt.Errorf("%s: lesson ID %q is already used by %s", path, lesson.ID, previous)
		}
		numbers[lesson.Number] = path
		ids[lesson.ID] = path
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

// loadFile decodes and validates one lesson YAML document.
func loadFile(files fs.FS, path string) (Lesson, error) {
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
