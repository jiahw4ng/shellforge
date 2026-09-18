package lessons

import "shellforge/internal/assertion"

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
	Title   string   `yaml:"title"`
	File    string   `yaml:"file"`
	Content Markdown `yaml:"-"`
}
