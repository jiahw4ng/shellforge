package assertion

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// Evaluate runs every assertion in the active lesson container and returns one
// result per assertion. Each command has a successful exit status for both a
// passing and failing condition; a non-nil Exec error therefore means Shellforge
// could not perform the check, rather than that the learner failed it.
func Evaluate(ctx context.Context, sandbox Executor, assertions []Assertion) []Result {
	return lo.Map(assertions, func(ass Assertion, _ int) Result {
		return evaluateSingleAssertion(ctx, sandbox, ass)
	})
}

func evaluateSingleAssertion(ctx context.Context, sandbox Executor, assertion Assertion) Result {
	script, message, args := assertionCommand(assertion)
	command := append([]string{"/bin/bash", "-c", script, "shellforge-assertion"}, args...)

	output, err := sandbox.Exec(ctx, "student", learnerWorkspace, command...)
	if err != nil {
		return Result{
			Assertion: assertion,
			Message:   fmt.Sprintf("Could not check whether %s: %v", message, err),
		}
	}

	return Result{Assertion: assertion, Passed: strings.TrimSpace(output) == "present", Message: message}
}

// assertionCommand returns the script, description, and arguments for one assertion.
func assertionCommand(assertion Assertion) (script string, description string, arguments []string) {
	switch assertion.Type {
	case AssertionTypeDirectoryExists:
		return directoryExistsScript,
			fmt.Sprintf("Directory %s", assertion.Path), []string{assertion.Path}
	case AssertionTypeFileExists:
		return fileExistsScript,
			fmt.Sprintf("File %s", assertion.Path), []string{assertion.Path}
	case AssertionTypeFileMode:
		return fileModeScript,
			fmt.Sprintf("File %s has mode %s", assertion.Path, assertion.Mode), []string{assertion.Path, assertion.Mode}
	case AssertionTypeFileContent:
		return fileContentScript,
			fmt.Sprintf("File %s with content %q", assertion.Path, assertion.Contains), []string{assertion.Path, assertion.Contains}
	case AssertionTypeCommandHistoryContains:
		return commandHistoryContainsScript,
			"Correct command executed", []string{commandHistoryFile, assertion.Contains}
	case AssertionTypeCurrentWorkingDirectory:
		return currentWorkingDirectoryScript,
			fmt.Sprintf("Current working directory is %s", assertion.Path), []string{workingDirectoryHistoryFile, assertion.Path}
	case AssertionTypeEnvironmentVariableExists:
		return environmentVariableExistsScript,
			fmt.Sprintf("Environment variable %s is exported", assertion.Name), []string{environmentSnapshotFile, assertion.Name}
	default:
		return "", "", nil
	}
}
