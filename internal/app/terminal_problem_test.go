package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"shellforge/internal/terminal"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func dockerStartFailure(reason container.Reason) error {
	return &terminal.TerminalStartError{
		Stage: "create lesson sandbox",
		Err: &container.ContainerUnavailableError{
			Reason: reason,
			Cause: &container.DockerCommandError{
				Arguments: []string{"info"},
				Err:       errors.New("exit status 1"),
				Output:    "Cannot connect to the Docker daemon at unix:///x.sock\n",
			},
		},
	}
}

func assertLearnerFacingProblem(t *testing.T, view string, wants ...string) {
	t.Helper()
	for _, want := range append(wants, "Then press Ctrl+D to go back and try again.", "Docker said: Cannot connect") {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q:\n%s", want, view)
		}
	}
	for _, internal := range []string{"create lesson sandbox", "docker lesson containers are unavailable", "exit status 1"} {
		if strings.Contains(view, internal) {
			t.Errorf("view leaks internal text %q:\n%s", internal, view)
		}
	}
}

func TestSandboxStartFailureShowsSummaryRemedyAndHowToLeave(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model.viewport = viewportState{width: 160, height: 30}
	model.term.err = dockerStartFailure(container.ReasonDockerUnreachable)

	view := ansi.Strip(model.View().Content)
	assertLearnerFacingProblem(t, view, "Docker isn't reachable.", "Start Docker with Linux containers enabled.")
}

func TestSandboxMissingImageShowsTheBuildCommand(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model.viewport = viewportState{width: 200, height: 30}
	model.term.err = dockerStartFailure(container.ReasonImageMissing)

	view := ansi.Strip(model.View().Content)
	assertLearnerFacingProblem(t, view, "image hasn't been built", container.ImageBuildCommand())
}

func TestLessonStartFailureShowsSummaryRemedyAndHowToLeave(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 220, height: 40}
	model.term.err = dockerStartFailure(container.ReasonDockerMissing)

	view := ansi.Strip(model.View().Content)
	assertLearnerFacingProblem(t, view, "couldn't find the Docker command")
}

func TestUnrecognisedStartFailureStillShowsItsText(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model.viewport = viewportState{width: 160, height: 30}
	model.term.err = &terminal.TerminalStartError{Stage: "start terminal emulator", Err: errors.New("pty failed")}

	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "start terminal emulator: pty failed") {
		t.Errorf("view does not show the unrecognised error's own text:\n%s", view)
	}
	if strings.Contains(view, "Then press") {
		t.Errorf("view shows learner advice for an error with no specific advice:\n%s", view)
	}
}

func TestTerminalOutputTakesPrecedenceOverTheProblemView(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model.viewport = viewportState{width: 160, height: 30}
	model.term.err = dockerStartFailure(container.ReasonDockerUnreachable)
	model.term.output = "bash: exited unexpectedly"

	view := ansi.Strip(model.View().Content)
	if !strings.Contains(view, "bash: exited unexpectedly") || strings.Contains(view, "Docker isn't reachable.") {
		t.Errorf("view should show the terminal's own output for an exit after start:\n%s", view)
	}
}
