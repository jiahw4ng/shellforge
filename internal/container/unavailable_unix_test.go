//go:build !windows

package container

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// useFakeDocker puts a stub docker script first on PATH, so the prerequisite
// checks run against known Docker output without needing a real engine.
func useFakeDocker(t *testing.T, script string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "docker")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o700); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
	t.Setenv("PATH", directory)
}

func unavailableReason(t *testing.T, err error) *ContainerUnavailableError {
	t.Helper()
	var unavailable *ContainerUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("CreateAndStart() error = %v, want an *UnavailableError", err)
	}
	if !errors.Is(err, ErrContainerUnavailable) {
		t.Errorf("errors.Is(ErrContainerUnavailable) = false for %v", err)
	}
	return unavailable
}

func TestCreateAndStartReportsMissingDockerCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := CreateAndStart(t.Context(), nil)
	unavailable := unavailableReason(t, err)
	if unavailable.Reason != ReasonDockerMissing {
		t.Fatalf("reason = %v, want %v", unavailable.Reason, ReasonDockerMissing)
	}
	if !errors.Is(err, exec.ErrNotFound) {
		t.Errorf("error = %v, want the LookPath cause preserved", err)
	}
}

func TestCreateAndStartKeepsDockerOutputWhenEngineUnreachable(t *testing.T) {
	useFakeDocker(t, `echo "Cannot connect to the Docker daemon at unix:///fake.sock" >&2
exit 1
`)

	_, err := CreateAndStart(t.Context(), nil)
	unavailable := unavailableReason(t, err)
	if unavailable.Reason != ReasonDockerUnreachable {
		t.Fatalf("reason = %v, want %v", unavailable.Reason, ReasonDockerUnreachable)
	}
	if !strings.Contains(err.Error(), "Cannot connect to the Docker daemon at unix:///fake.sock") {
		t.Errorf("error = %q, want Docker's own message instead of a guess", err)
	}
	var command *DockerCommandError
	if !errors.As(err, &command) || !strings.Contains(command.Output, "Cannot connect to the Docker daemon") {
		t.Errorf("errors.As(*DockerCommandError) = %#v, want Docker's output kept for the learner message", command)
	}
}

func TestCreateAndStartKeepsDockerOutputWhenImageIsMissing(t *testing.T) {
	useFakeDocker(t, `case "$1" in
  info) exit 0 ;;
  *) echo "Error: No such image: fake" >&2; exit 1 ;;
esac
`)

	_, err := CreateAndStart(t.Context(), nil)
	unavailable := unavailableReason(t, err)
	if unavailable.Reason != ReasonImageMissing {
		t.Fatalf("reason = %v, want %v", unavailable.Reason, ReasonImageMissing)
	}
	if !strings.Contains(err.Error(), "No such image: fake") {
		t.Errorf("error = %q, want Docker's own message", err)
	}
}
