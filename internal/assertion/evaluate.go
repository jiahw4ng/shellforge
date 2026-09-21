package assertion

import (
	"context"
	"fmt"
	"strings"
)

const learnerWorkspace = "/home/student/workspace"

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
	script, checkFor, args, ok := assertionCommand(assertion)

	// check for unsupported assertion types
	// shld not happen technically
	if !ok {
		return Result{
			Assertion: assertion,
			Message:   fmt.Sprintf("Unsupported assertion type %q.", assertion.Type),
		}
	}

	// get command to execute
	command := append([]string{"/bin/bash", "-c", script, "shellforge-assertion"}, args...)

	// run the command in the container
	output, err := sandbox.Exec(ctx, "student", learnerWorkspace, command...)
	// check for execution errors
	if err != nil {
		return Result{
			Assertion: assertion,
			Message:   fmt.Sprintf("Could not check whether %s: %v", checkFor, err),
		}
	}

	// check for assertion result
	passed := strings.TrimSpace(output) == "present"
	var message string
	if passed {
		message = fmt.Sprintf("%s exists.", checkFor)
	} else {
		message = fmt.Sprintf("%s does not exist.", checkFor)
	}
	return Result{Assertion: assertion, Passed: passed, Message: message}
}

// assertionCommand returns the script, description, and arguments for one assertion.
func assertionCommand(assertion Assertion) (script string, description string, arguments []string, ok bool) {
	switch assertion.Type {
	case AssertionTypeDirectoryExists:
		return DirectoryExistsScript,
			fmt.Sprintf("Directory %q", assertion.Path), []string{assertion.Path}, true
	case AssertionTypeFileExists:
		return FileExistsScript,
			fmt.Sprintf("File %q", assertion.Path), []string{assertion.Path}, true
	case AssertionTypeFileContent, AssertionTypeFileContains:
		return FileContentScript,
			fmt.Sprintf("File %q with content %q", assertion.Path, assertion.Contains), []string{assertion.Path, assertion.Contains}, true
	default:
		return "", "", nil, false
	}
}
