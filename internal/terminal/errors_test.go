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

func startFailure(reason container.Reason, cause error) error {
	return &TerminalStartError{
		Stage: createSandboxStage,
		Err:   &container.ContainerUnavailableError{Reason: reason, Cause: cause},
	}
}

func dockerCommandFailure(output string) error {
	return &container.DockerCommandError{Arguments: []string{"info"}, Err: errors.New("exit status 1"), Output: output}
}

func TestDescribeExplainsEachUnavailableReason(t *testing.T) {
	cause := dockerCommandFailure("\n  Cannot connect to the Docker daemon at unix:///x.sock\nsecond line\n")
	tests := []struct {
		name       string
		reason     container.Reason
		wantInSumm string
		wantInFix  string
	}{
		{"docker missing", container.ReasonDockerMissing, "couldn't find the Docker command", "PATH"},
		{"docker unreachable", container.ReasonDockerUnreachable, "isn't reachable", "Linux containers"},
		{"image missing", container.ReasonImageMissing, "couldn't download the lesson sandbox image", container.ImageBuildCommand()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, ok := Describe(startFailure(test.reason, cause))
			if !ok {
				t.Fatal("Describe() ok = false, want a learner message")
			}
			if !strings.Contains(message.Summary, test.wantInSumm) {
				t.Errorf("Summary = %q, want it to contain %q", message.Summary, test.wantInSumm)
			}
			if !strings.Contains(message.Remedy, test.wantInFix) {
				t.Errorf("Remedy = %q, want it to contain %q", message.Remedy, test.wantInFix)
			}
			if want := "Docker said: Cannot connect to the Docker daemon at unix:///x.sock"; message.Detail != want {
				t.Errorf("Detail = %q, want the first non-empty Docker line %q", message.Detail, want)
			}
			everything := message.Summary + message.Remedy + message.Detail
			for _, internal := range []string{createSandboxStage, container.ErrContainerUnavailable.Error(), "exit status"} {
				if strings.Contains(everything, internal) {
					t.Errorf("learner message %q leaks internal text %q", everything, internal)
				}
			}
		})
	}
}

func TestDescribeLeavesOtherErrorsToTheCaller(t *testing.T) {
	tests := map[string]error{
		"plain error":          errors.New("boom"),
		"unrelated start":      &TerminalStartError{Stage: startEmulatorStage, Err: errors.New("pty failed")},
		"unknown reason":       startFailure(container.Reason(99), errors.New("boom")),
		"nil":                  nil,
		"invalid start result": NewInvalidStartResultError(),
	}
	for name, err := range tests {
		t.Run(name, func(t *testing.T) {
			if message, ok := Describe(err); ok {
				t.Fatalf("Describe() = %#v, true; want ok=false so the raw text is shown", message)
			}
		})
	}
}

func TestDescribeOmitsDetailWithoutDockerOutput(t *testing.T) {
	for name, cause := range map[string]error{
		"not a docker command error": errors.New("exec: docker: not found"),
		"empty output":               dockerCommandFailure("  \n\n"),
	} {
		t.Run(name, func(t *testing.T) {
			message, ok := Describe(startFailure(container.ReasonDockerMissing, cause))
			if !ok || message.Detail != "" {
				t.Fatalf("Describe() = %#v, %v; want a message with no detail", message, ok)
			}
		})
	}
}

func TestDescribeShortensVeryLongDockerLines(t *testing.T) {
	message, _ := Describe(startFailure(container.ReasonDockerUnreachable, dockerCommandFailure(strings.Repeat("x", 500))))

	wantRunes := len("Docker said: ") + maxDetailRunes + len("…")
	if got := len([]rune(message.Detail)); got > wantRunes || !strings.HasSuffix(message.Detail, "…") {
		t.Fatalf("Detail has %d runes (%q...), want at most %d ending in an ellipsis", got, message.Detail[:20], wantRunes)
	}
}
