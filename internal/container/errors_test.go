package container

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestUnavailableErrorMatchesCategoryAndKeepsCause(t *testing.T) {
	cause := errors.New("Cannot connect to the Docker daemon")
	for _, reason := range []Reason{ReasonDockerMissing, ReasonDockerUnreachable, ReasonImageMissing} {
		t.Run(reason.String(), func(t *testing.T) {
			err := error(&ContainerUnavailableError{Reason: reason, Cause: cause})
			wrapped := fmt.Errorf("outer: %w", err)

			if !errors.Is(wrapped, ErrContainerUnavailable) {
				t.Error("errors.Is(ErrContainerUnavailable) = false, want the whole category to match")
			}
			if !errors.Is(wrapped, cause) {
				t.Error("errors.Is(cause) = false, want the original cause reachable")
			}
			var unavailable *ContainerUnavailableError
			if !errors.As(wrapped, &unavailable) || unavailable.Reason != reason {
				t.Errorf("errors.As() = %v, reason %v; want reason %v", unavailable, unavailable, reason)
			}
			for _, want := range []string{reason.String(), cause.Error()} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Error() = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestUnavailableErrorDoesNotRepeatTheSentinelText(t *testing.T) {
	err := &ContainerUnavailableError{Reason: ReasonDockerUnreachable, Cause: errors.New("boom")}
	if strings.Contains(err.Error(), ErrContainerUnavailable.Error()) {
		t.Fatalf("Error() = %q, want the sentinel to stay a category, not display text", err)
	}
}

func TestOtherErrorsDoNotMatchTheCategory(t *testing.T) {
	if errors.Is(errors.New("docker create failed"), ErrContainerUnavailable) {
		t.Fatal("an unrelated error matched ErrContainerUnavailable")
	}
}

func TestImageMissingReasonNamesTheImageAndBuildCommand(t *testing.T) {
	text := ReasonImageMissing.String()
	for _, want := range []string{image, "docker build", "docker/lesson.Dockerfile"} {
		if !strings.Contains(text, want) {
			t.Errorf("ReasonImageMissing = %q, want it to contain %q", text, want)
		}
	}
}
