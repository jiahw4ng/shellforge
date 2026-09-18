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
	Content        string                `yaml:"content"`
	Hints          []string              `yaml:"hints"`
	Setup          string                `yaml:"setup"`
	Assertions     []assertion.Assertion `yaml:"assertions"`
	SuccessMessage string                `yaml:"success_message"`
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
	case l.Content == "":
		return fmt.Errorf("missing content")
	}
	return nil
}
