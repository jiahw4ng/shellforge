package assertion

import (
	"context"
	"fmt"
	"strings"
)

// Evaluate runs every assertion in the active lesson container and returns one
// result per assertion. Each command has a successful exit status for both a
// passing and failing condition; a non-nil Exec error therefore means Shellforge
// could not perform the check, rather than that the learner failed it.
func Evaluate(ctx context.Context, sandbox Executor, assertions []Assertion) []Result {
	results := make([]Result, 0, len(assertions))
	for _, assertion := range assertions {
		results = append(results, evaluateSingleAssertion(ctx, sandbox, assertion))
	}
	return results
}

func evaluateSingleAssertion(ctx context.Context, sandbox Executor, assertion Assertion) Result {
	script, message, args := assertionCommand(assertion)

	// get command to execute
	command := append([]string{"/bin/bash", "-c", script, "shellforge-assertion"}, args...)

	// run the command in the container
	output, err := sandbox.Exec(ctx, "student", learnerWorkspace, command...)
	// check for execution errors
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
		return DirectoryExistsScript,
			fmt.Sprintf("Directory %s", assertion.Path), []string{assertion.Path}
	case AssertionTypeFileExists:
		return FileExistsScript,
			fmt.Sprintf("File %s", assertion.Path), []string{assertion.Path}
	case AssertionTypeFileContent:
		return FileContentScript,
			fmt.Sprintf("File %s with content %q", assertion.Path, assertion.Contains), []string{assertion.Path, assertion.Contains}
	case AssertionTypeCommandHistoryContains:
		return CommandHistoryContainsScript,
			"Correct command executed", []string{commandHistoryFile, assertion.Contains}
	case AssertionTypeCurrentWorkingDirectory:
		return CurrentWorkingDirectoryScript,
			fmt.Sprintf("Current working directory is %s", assertion.Path), []string{workingDirectoryHistoryFile, assertion.Path}
	case AssertionTypeEnvironmentVariableExists:
		return EnvironmentVariableExistsScript,
			fmt.Sprintf("Environment variable %s is exported", assertion.Name), []string{environmentSnapshotFile, assertion.Name}
	default:
		return "", "", nil
	}
}

func HasPassedAllAssertions(results []Result) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}
