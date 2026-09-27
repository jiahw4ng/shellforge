package assertion

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	if results[0].Message != `File notes/idea.txt` {
		t.Fatalf("message = %q, want file check label", results[0].Message)
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

func TestEvaluateChecksCommandHistory(t *testing.T) {
	executor := &fakeExecutor{output: "present"}
	assertion := Assertion{Type: AssertionTypeCommandHistoryContains, Contains: "cd project"}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("Evaluate() = %#v, want one passing result", results)
	}

	command := strings.Join(executor.command, "\x00")
	if !strings.Contains(command, "/home/student/.shellforge-history") {
		t.Fatalf("command = %#v, want Shellforge history path", executor.command)
	}
	if !strings.Contains(command, "cd project") {
		t.Fatalf("command = %#v, want expected command text", executor.command)
	}
}

func TestEvaluateChecksCurrentWorkingDirectory(t *testing.T) {
	executor := &fakeExecutor{output: "present"}
	assertion := Assertion{
		Type: AssertionTypeCurrentWorkingDirectory,
		Path: "/home/student/workspace/project/docs",
	}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("Evaluate() = %#v, want one passing result", results)
	}

	command := strings.Join(executor.command, "\x00")
	for _, required := range []string{
		"tail -n 1",
		"/home/student/.shellforge-working-directories",
		"/home/student/workspace/project/docs",
	} {
		if !strings.Contains(command, required) {
			t.Errorf("command = %#v, want %q", executor.command, required)
		}
	}
}

func TestEvaluateReportsCurrentWorkingDirectoryMismatch(t *testing.T) {
	executor := &fakeExecutor{output: "missing"}
	assertion := Assertion{
		Type: AssertionTypeCurrentWorkingDirectory,
		Path: "/home/student/workspace/project/docs",
	}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if results[0].Passed {
		t.Fatalf("result = %#v, want failed assertion", results[0])
	}
	if results[0].Message != "Current working directory is /home/student/workspace/project/docs" {
		t.Fatalf("message = %q, want current-working-directory label", results[0].Message)
	}
}

func TestEvaluateChecksEnvironmentVariable(t *testing.T) {
	executor := &fakeExecutor{output: "present"}
	assertion := Assertion{Type: AssertionTypeEnvironmentVariableExists, Name: "EDITOR"}

	results := Evaluate(context.Background(), executor, []Assertion{assertion})
	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("Evaluate() = %#v, want one passing result", results)
	}

	command := strings.Join(executor.command, "\x00")
	for _, required := range []string{
		"read -r -d '' entry",
		"/home/student/.shellforge-environment",
		"EDITOR",
	} {
		if !strings.Contains(command, required) {
			t.Errorf("command = %#v, want %q", executor.command, required)
		}
	}
}

func TestEnvironmentVariableExistsScriptReadsNULDelimitedEntries(t *testing.T) {
	snapshotPath := filepath.Join(t.TempDir(), "environment")
	if err := os.WriteFile(snapshotPath, []byte("PATH=/usr/bin\x00MESSAGE=EDITOR=inside-a-value\x00"), 0o600); err != nil {
		t.Fatal(err)
	}

	for variableName, want := range map[string]string{
		"PATH":   "present",
		"EDITOR": "missing",
	} {
		t.Run(variableName, func(t *testing.T) {
			output, err := exec.Command("/bin/bash", "-c", EnvironmentVariableExistsScript, "shellforge-assertion", snapshotPath, variableName).Output()
			if err != nil {
				t.Fatalf("environment assertion script error = %v", err)
			}
			if got := strings.TrimSpace(string(output)); got != want {
				t.Fatalf("environment assertion script output = %q, want %q", got, want)
			}
		})
	}
}
