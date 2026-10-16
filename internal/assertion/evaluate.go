package assertion

import (
	"context"
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

	// get command to execute
	command := append([]string{bashExecutable, "-c", script, assertionCommandName}, args...)

	// run the command in the container
	output, err := sandbox.Exec(ctx, learnerUser, learnerWorkspace, command...)
	// check for execution errors
	if err != nil {
		return Result{
			Assertion: assertion,
			Message:   message,
			Err:       err,
		}
	}

	return Result{Assertion: assertion, Passed: strings.TrimSpace(output) == assertionPassedOutput, Message: message}
}

// assertionCommand returns the script, description, and arguments for one assertion.
func assertionCommand(assertion Assertion) (script string, description string, arguments []string) {
	switch assertion.Type {
	case AssertionTypeDirectoryExists:
		return DirectoryExistsScript, assertion.Task(), []string{assertion.Path}
	case AssertionTypeFileExists:
		return FileExistsScript, assertion.Task(), []string{assertion.Path}
	case AssertionTypeFileMode:
		return FileModeScript, assertion.Task(), []string{assertion.Path, assertion.Mode}
	case AssertionTypeFileContent:
		return FileContentScript, assertion.Task(), []string{assertion.Path, assertion.Contains}
	case AssertionTypeCommandHistoryContains:
		return CommandHistoryContainsScript, assertion.Task(), []string{commandHistoryFile, assertion.Contains}
	case AssertionTypeCurrentWorkingDirectory:
		return CurrentWorkingDirectoryScript, assertion.Task(), []string{workingDirectoryHistoryFile, assertion.Path}
	case AssertionTypeEnvironmentVariableExists:
		return EnvironmentVariableExistsScript, assertion.Task(), []string{environmentSnapshotFile, assertion.Name}
	default:
		return "", "", nil
	}
}
