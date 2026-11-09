package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func startingSandboxState() State {
	state := State{
		nav:      navigationState{screen: config.SandboxTerminalScreen},
		viewport: viewportState{width: 120, height: 30},
	}
	state.term.beginAttempt()
	state.term.startEvents = make(chan tea.Msg)
	return state
}

func TestStartProgressAccumulatesStepsInOrderAndWaitsForTheNext(t *testing.T) {
	model := startingSandboxState()
	steps := []container.Step{container.StepCheckingDocker, container.StepCheckingImage}

	for _, step := range steps {
		updated, command := model.Update(termStartProgressMsg{step: step, gen: model.term.gen})
		model = updated.(State)
		if command == nil {
			t.Fatalf("no command after %v, want a wait for the next startup event", step)
		}
	}
	if !slices.Equal(model.term.startSteps, steps) {
		t.Fatalf("startSteps = %v, want %v", model.term.startSteps, steps)
	}
}

func TestStartProgressFromAnOlderAttemptIsIgnored(t *testing.T) {
	model := startingSandboxState()

	updated, command := model.Update(termStartProgressMsg{step: container.StepCheckingDocker, gen: model.term.gen - 1})
	model = updated.(State)
	if len(model.term.startSteps) != 0 || command != nil {
		t.Fatalf("stale progress changed state: steps=%v command=%v", model.term.startSteps, command != nil)
	}
}

func TestNewAttemptForgetsEarlierStartupSteps(t *testing.T) {
	model := startingSandboxState()
	model.term.startSteps = []container.Step{container.StepCheckingDocker}

	model.term.beginAttempt()
	if len(model.term.startSteps) != 0 {
		t.Fatalf("startSteps = %v after a new attempt, want none", model.term.startSteps)
	}
}

func TestStartupLinesShowInTheSandboxViewWithTheSpinnerOnTheLastLine(t *testing.T) {
	model := startingSandboxState()
	model.term.startSteps = []container.Step{
		container.StepCheckingDocker,
		container.StepCheckingImage,
		container.StepPullingImage,
		container.StepStartingContainer,
	}

	lines := strings.Split(ansi.Strip(model.View().Content), "\n")
	want := []string{
		">_ Sandboxed Bash",
		"Checking that Docker is running",
		"Checking if sandbox image exists",
		"Image does not exist, pulling from registry",
		"Image pulled, starting container",
	}
	// The pane sits inside a border, so find each line by its text and require
	// them to appear one directly after another.
	first := slices.IndexFunc(lines, func(line string) bool { return strings.Contains(line, want[0]) })
	if first < 0 || first+len(want) > len(lines) {
		t.Fatalf("view has no %q header:\n%s", want[0], strings.Join(lines, "\n"))
	}
	for index, text := range want {
		if !strings.Contains(lines[first+index], text) {
			t.Fatalf("line %d = %q, want it to contain %q\nview:\n%s", first+index, lines[first+index], text, strings.Join(lines, "\n"))
		}
	}
	spinner := ansi.Strip(model.term.spinner.View())
	if strings.Contains(lines[first+len(want)-2], spinner) {
		t.Errorf("spinner on a finished line: %q", lines[first+len(want)-2])
	}
	if !strings.Contains(lines[first+len(want)-1], spinner) {
		t.Errorf("last line %q has no spinner %q", lines[first+len(want)-1], spinner)
	}
}

func TestStartupLinesDisappearWhenStartFails(t *testing.T) {
	model := startingSandboxState()
	model.term.startSteps = []container.Step{container.StepCheckingDocker, container.StepCheckingImage}

	model = updateModel(t, model, termStartedMsg{err: errors.New("docker exploded"), gen: model.term.gen})

	view := ansi.Strip(model.View().Content)
	if strings.Contains(view, "Checking") {
		t.Errorf("view still shows progress lines after a failure:\n%s", view)
	}
	if !strings.Contains(view, "docker exploded") {
		t.Errorf("view = %q, want the error screen", view)
	}
}
