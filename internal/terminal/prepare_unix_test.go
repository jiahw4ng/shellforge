//go:build !windows

package terminal

import (
	"os"
	"path/filepath"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"slices"
	"testing"
)

// useSucceedingDocker puts a stub docker that accepts every command first on PATH.
func useSucceedingDocker(t *testing.T) {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "docker"), []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
	t.Setenv("PATH", directory)
}

func TestPrepareSandboxReportsPreparingLessonOnlyWhenThereIsSetup(t *testing.T) {
	tests := []struct {
		name   string
		lesson *lessons.Lesson
		want   []container.Step
	}{
		{"standalone sandbox", nil, []container.Step{container.StepCheckingDocker, container.StepCheckingImage, container.StepStartingContainer}},
		{"lesson without setup", &lessons.Lesson{}, []container.Step{container.StepCheckingDocker, container.StepCheckingImage, container.StepStartingContainer}},
		{"lesson with setup", &lessons.Lesson{Setup: "touch notes.txt"}, []container.Step{container.StepCheckingDocker, container.StepCheckingImage, container.StepStartingContainer, container.StepPreparingLesson}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useSucceedingDocker(t)
			var got []container.Step
			if _, err := prepareSandbox(t.Context(), test.lesson, func(step container.Step) { got = append(got, step) }); err != nil {
				t.Fatalf("prepareSandbox() error = %v", err)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("steps = %v, want %v", got, test.want)
			}
		})
	}
}
