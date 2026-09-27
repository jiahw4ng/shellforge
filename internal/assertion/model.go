// Package assertion evaluates the observable outcomes defined by lessons.
package assertion

import "context"

type AssertionKind string

const (
	// checks whether a directory exists at the given path
	AssertionTypeDirectoryExists AssertionKind = "directory_exists"
	// checks whether a file exists at the given path
	AssertionTypeFileExists AssertionKind = "file_exists"
	// checks whether a file exists at the given path and contains the given string
	AssertionTypeFileContent AssertionKind = "file_content"
	// checks whether the command history contains the given string
	AssertionTypeCommandHistoryContains AssertionKind = "command_history_contains"
	// checks whether the learner's latest interactive working directory matches the given path
	AssertionTypeCurrentWorkingDirectory AssertionKind = "cwd"
)

// Assertion describes one future state check run separately inside the lesson
// container. It verifies results rather than requiring one exact command.
type Assertion struct {
	Type     AssertionKind `yaml:"type"`
	Path     string        `yaml:"path,omitempty"`
	Contains string        `yaml:"contains,omitempty"`
}

// Result is the outcome of one assertion check run inside the lesson container.
type Result struct {
	Assertion Assertion
	Passed    bool
	Message   string
}

// Executor runs a command in the active lesson container. Container satisfies
// this interface, while tests can use a small fake without requiring Docker.
type Executor interface {
	Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error)
}
