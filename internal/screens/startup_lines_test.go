package screens

import (
	"shellforge/internal/container"
	"slices"
	"testing"
)

func TestStartupLinesWordTheContainerStepByWhetherTheImageWasPulled(t *testing.T) {
	tests := []struct {
		name  string
		steps []container.Step
		want  []string
	}{
		{
			"image found",
			[]container.Step{container.StepCheckingDocker, container.StepCheckingImage, container.StepStartingContainer},
			[]string{"Checking that Docker is running", "Checking if sandbox image exists", "Image found, starting container"},
		},
		{
			"image pulled",
			[]container.Step{container.StepCheckingImage, container.StepPullingImage, container.StepStartingContainer},
			[]string{"Checking if sandbox image exists", "Image does not exist, pulling from registry", "Image pulled, starting container"},
		},
		{
			"lesson with setup",
			[]container.Step{container.StepStartingContainer, container.StepPreparingLesson, container.StepConnectingShell},
			[]string{"Image found, starting container", "Preparing lesson", "Connecting to the shell"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StartupLines(test.steps); !slices.Equal(got, test.want) {
				t.Errorf("StartupLines() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTerminalStartWithoutStepsFallsBackToTheGenericLine(t *testing.T) {
	if got := TerminalStart(nil, nil, ""); got == "" {
		t.Fatal("TerminalStart() is empty, want the generic starting line")
	}
}
