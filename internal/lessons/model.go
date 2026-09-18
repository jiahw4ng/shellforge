package lessons

import (
	"fmt"
	"shellforge/internal/assertion"
)

// Lesson is the complete, version-controlled definition of one Shellforge
// exercise. Setup is trusted project-authored shell code run in the container
// before the learner's Bash session starts.
type Lesson struct {
	ID             string                `yaml:"id"`
	Number         int                   `yaml:"number"`
	Title          string                `yaml:"title"`
	Description    string                `yaml:"description"`
	Pages          []Page                `yaml:"pages"`
	Hints          []string              `yaml:"hints"`
	Setup          string                `yaml:"setup"`
	Assertions     []assertion.Assertion `yaml:"assertions"`
	SuccessMessage string                `yaml:"success_message"`
}

// Markdown represents a string of Markdown content.
// It is a distinct type to avoid accidental assignment of arbitrary strings to lesson content fields.
type Markdown string

// Page is one focused unit of lesson instruction shown beside the terminal.
type Page struct {
	Title   string `yaml:"title"`
	Content Markdown `yaml:"content"`
}

// validate rejects incomplete lesson definitions before they reach the UI or
// container setup runner.
func (l Lesson) validate() error {
	switch {
	case l.ID == "":
		return fmt.Errorf("missing id")
	case l.Number < 1:
		return fmt.Errorf("number must be at least 1")
	case l.Title == "":
		return fmt.Errorf("missing title")
	case len(l.Pages) == 0:
		return fmt.Errorf("must contain at least one page")
	}

	for index, page := range l.Pages {
		if page.Title == "" {
			return fmt.Errorf("page %d: missing title", index+1)
		}
		if page.Content == "" {
			return fmt.Errorf("page %d: missing content", index+1)
		}
	}
	return nil
}
