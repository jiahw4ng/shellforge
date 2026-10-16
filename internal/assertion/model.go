// Package assertion evaluates the observable outcomes defined by lessons.
package assertion

import (
	"context"
	"fmt"
)

// Assertion describes one future state check run separately inside the lesson
// container. It verifies results rather than requiring one exact command.
type Assertion struct {
	Type     AssertionKind `yaml:"type"`
	Path     string        `yaml:"path,omitempty"`
	Contains string        `yaml:"contains,omitempty"`
	Mode     string        `yaml:"mode,omitempty"`
	Name     string        `yaml:"name,omitempty"`
}

// Task returns a learner-facing instruction for the observable outcome.
func (check Assertion) Task() string {
	switch check.Type {
	case AssertionTypeDirectoryExists:
		return fmt.Sprintf("Create directory %s", check.Path)
	case AssertionTypeFileExists:
		return fmt.Sprintf("Create file %s", check.Path)
	case AssertionTypeFileMode:
		return fmt.Sprintf("Set %s permissions to %s", check.Path, check.Mode)
	case AssertionTypeFileContent:
		return fmt.Sprintf("Add %q to %s", check.Contains, check.Path)
	case AssertionTypeCommandHistoryContains:
		return fmt.Sprintf("Run a command containing %q", check.Contains)
	case AssertionTypeCurrentWorkingDirectory:
		return fmt.Sprintf("Change directory to %s", check.Path)
	case AssertionTypeEnvironmentVariableExists:
		return fmt.Sprintf("Export environment variable %s", check.Name)
	default:
		return ""
	}
}

// Result is the outcome of one assertion check run inside the lesson container.
// Err is non-nil only when Shellforge could not evaluate the assertion; an
// unmet assertion has Passed set to false and a nil Err.
type Result struct {
	Assertion Assertion
	Passed    bool
	Message   string
	Err       error
}

// Executor runs a command in the active lesson container. container.Container satisfies
// this interface, while tests can use a small fake without requiring Docker.
type Executor interface {
	Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error)
}
