package screens

import (
	"errors"
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestAssertionResultsRendersEachOutcomeThenOneSummary(t *testing.T) {
	passed := assertion.Result{Passed: true, Message: "Create file a"}
	pending := assertion.Result{Message: "Create file b"}
	failed := assertion.Result{Message: "Create file c", Err: errors.New("sandbox stopped")}

	tests := []struct {
		name    string
		results []assertion.Result
		want    []string
	}{
		{"none", nil, []string{"✓ There are no progress checks for this page."}},
		{"all passed", []assertion.Result{passed, passed}, []string{
			"✓ Create file a", "✓ Create file a", "Well done.", "All checks passed!",
		}},
		{"some pending", []assertion.Result{passed, pending}, []string{
			"✓ Create file a", "○ Create file b", "Press F12 to check progress.",
		}},
		{"execution error outranks pending results", []assertion.Result{pending, failed, passed}, []string{
			"○ Create file b", "✗ Could not check Create file c: sandbox stopped", "✓ Create file a",
			"Unable to run some checks!",
		}},
		{"execution error alone", []assertion.Result{failed}, []string{
			"✗ Could not check Create file c: sandbox stopped", "Unable to run some checks!",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ansi.Strip(assertionResults(LessonRenderParams{
				Lesson:           &lessons.Lesson{SuccessMessage: "Well done."},
				AssertionResults: test.results,
			}))
			if want := strings.Join(test.want, "\n"); got != want {
				t.Fatalf("assertionResults() =\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func TestAssertionStatusBodyPrefersTerminalFailureThenCheckingThenResults(t *testing.T) {
	lesson := &lessons.Lesson{
		SuccessMessage: "Well done.",
		Assertions:     []assertion.Assertion{{Type: assertion.AssertionTypeDirectoryExists, Path: "/home/student/workspace/a"}},
	}
	results := []assertion.Result{{Passed: true, Message: "Create a"}}

	tests := []struct {
		name string
		p    LessonRenderParams
		want string
	}{
		{"terminal failure wins", LessonRenderParams{Lesson: lesson, TerminalError: errors.New("x"), AssertionsChecking: true, AssertionsChecked: true, AssertionResults: results}, "Progress checks are not available"},
		{"checking beats checked", LessonRenderParams{Lesson: lesson, AssertionsChecking: true, AssertionsChecked: true, AssertionResults: results}, "Checking Progress..."},
		{"checked shows results", LessonRenderParams{Lesson: lesson, AssertionsChecked: true, AssertionResults: results}, "All checks passed!"},
		{"otherwise prompts", LessonRenderParams{Lesson: lesson}, "Press F12 to check progress."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.p.AssertionsExpanded = true // a collapsed disclosure hides the body under test
			if got := ansi.Strip(assertionStatus(test.p, 80)); !strings.Contains(got, test.want) {
				t.Fatalf("assertionStatus() = %q, want it to contain %q", got, test.want)
			}
		})
	}
}
