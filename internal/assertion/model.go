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
func (a Assertion) Task() string {
	switch a.Type {
	case AssertionTypeDirectoryExists:
		return fmt.Sprintf(createDirectoryInstruction, a.Path)
	case AssertionTypeFileExists:
		return fmt.Sprintf(createFileInstruction, a.Path)
	case AssertionTypeFileMode:
		return fmt.Sprintf(setFileModeInstruction, a.Path, a.Mode)
	case AssertionTypeFileContent:
		return fmt.Sprintf(addFileContentInstruction, a.Contains, a.Path)
	case AssertionTypeCommandHistoryContains:
		return fmt.Sprintf(runCommandInstruction, a.Contains)
	case AssertionTypeCurrentWorkingDirectory:
		return fmt.Sprintf(changeDirectoryInstruction, a.Path)
	case AssertionTypeEnvironmentVariableExists:
		return fmt.Sprintf(exportEnvVarInstruction, a.Name)
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
