package terminal

import (
	"errors"
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
