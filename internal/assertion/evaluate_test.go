package assertion

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeExecutor struct {
	output  string
	err     error
	command []string
}

func (f *fakeExecutor) Exec(_ context.Context, _ string, _ string, command ...string) (string, error) {
	f.command = command
	return f.output, f.err
}

func TestEvaluateChecksDirectoryInsideLearnerWorkspace(t *testing.T) {
	executor := &fakeExecutor{output: "present"}
	assertion := Assertion{Type: AssertionTypeDirectoryExists, Path: "notes"}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("Evaluate() = %#v, want one passing result", results)
	}
	if !strings.Contains(strings.Join(executor.command, "\x00"), "-d") {
		t.Fatalf("command = %#v, want directory test", executor.command)
	}
}

func TestEvaluateReportsMissingFileWithoutTreatingItAsExecutionError(t *testing.T) {
	executor := &fakeExecutor{output: "missing"}
	assertion := Assertion{Type: AssertionTypeFileExists, Path: "notes/idea.txt"}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if results[0].Passed {
		t.Fatalf("result = %#v, want failed assertion", results[0])
	}
	if !strings.Contains(results[0].Message, "does not exist") {
		t.Fatalf("message = %q, want missing-file message", results[0].Message)
	}
}

func TestEvaluateReportsContainerExecutionFailure(t *testing.T) {
	executor := &fakeExecutor{err: errors.New("Docker stopped")}
	assertion := Assertion{Type: AssertionTypeFileExists, Path: "notes/idea.txt"}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if results[0].Passed || !strings.Contains(results[0].Message, "Could not check") {
		t.Fatalf("result = %#v, want execution-failure message", results[0])
	}
}
