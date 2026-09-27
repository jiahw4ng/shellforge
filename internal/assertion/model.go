// Package assertion evaluates the observable outcomes defined by lessons.
package assertion

import "context"

// Assertion describes one future state check run separately inside the lesson
// container. It verifies results rather than requiring one exact command.
type Assertion struct {
	Type     AssertionKind `yaml:"type"`
	Path     string        `yaml:"path,omitempty"`
	Contains string        `yaml:"contains,omitempty"`
	Mode     string        `yaml:"mode,omitempty"`
	Name     string        `yaml:"name,omitempty"`
}

// Result is the outcome of one assertion check run inside the lesson container.
type Result struct {
	Assertion Assertion
	Passed    bool
	Message   string
}

// Executor runs a command in the active lesson container. container.Container satisfies
// this interface, while tests can use a small fake without requiring Docker.
type Executor interface {
	Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error)
}
