package terminal

import (
	"errors"
	"shellforge/internal/container"
	"strings"
	"testing"
)

func TestStartErrorPreservesStageAndCause(t *testing.T) {
	cause := errors.New("Docker is unavailable")
	err := &TerminalStartError{Stage: "create lesson sandbox", Err: cause}

	if !strings.Contains(err.Error(), "create lesson sandbox") {
		t.Fatalf("Error() = %q, want stage", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is() did not find the original cause")
	}
}

func TestStartErrorExposesContainerUnavailableReason(t *testing.T) {
	cause := errors.New("Cannot connect to the Docker daemon")
	err := &TerminalStartError{
		Stage: createSandboxStage,
		Err:   &container.ContainerUnavailableError{Reason: container.ReasonDockerUnreachable, Cause: cause},
	}

	if !errors.Is(err, container.ErrContainerUnavailable) {
		t.Error("errors.Is(ErrContainerUnavailable) = false through TerminalStartError")
	}
	var unavailable *container.ContainerUnavailableError
	if !errors.As(err, &unavailable) || unavailable.Reason != container.ReasonDockerUnreachable {
		t.Errorf("errors.As() = %v, want the unreachable reason", unavailable)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is(cause) = false, want Docker's original error reachable")
	}
}
