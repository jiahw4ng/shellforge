//go:build !windows

package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// useRecordingDocker installs a stub docker that logs every call and fails the
// given subcommand (or none when failOn is empty). It returns the log's path.
func useRecordingDocker(t *testing.T, failOn string) string {
	t.Helper()
	log := filepath.Join(t.TempDir(), "calls.log")
	failing := failOn
	if failing == "" {
		failing = "no-such-subcommand"
	}
	useFakeDocker(t, fmt.Sprintf(`echo "$*" >> %q
case "$1" in
  %s) echo "boom" >&2; exit 1 ;;
esac
exit 0
`, log, failing))
	return log
}

// recordedCalls returns the logged docker invocations, one "subcommand args" per entry.
func recordedCalls(t *testing.T, log string) []string {
	t.Helper()
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("read docker call log: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func subcommands(calls []string) []string {
	names := make([]string, len(calls))
	for index, call := range calls {
		names[index], _, _ = strings.Cut(call, " ")
	}
	return names
}

func TestCreateAndStartChecksPrerequisitesThenCreatesAndStarts(t *testing.T) {
	log := useRecordingDocker(t, "")

	container, err := CreateAndStart(t.Context())
	if err != nil {
		t.Fatalf("CreateAndStart() error = %v", err)
	}
	if !strings.HasPrefix(container.Name, containerNamePrefix) {
		t.Errorf("container name = %q, want prefix %q", container.Name, containerNamePrefix)
	}

	calls := recordedCalls(t, log)
	if got, want := strings.Join(subcommands(calls), " "), "info image create start"; got != want {
		t.Fatalf("docker calls = %q, want %q (prerequisites first, never a removal)", got, want)
	}
	if want := "start " + container.Name; calls[3] != want {
		t.Errorf("start call = %q, want %q", calls[3], want)
	}
}

func TestCreateAndStartStopsWhenCreateFails(t *testing.T) {
	log := useRecordingDocker(t, "create")

	if _, err := CreateAndStart(t.Context()); err == nil {
		t.Fatal("CreateAndStart() error = nil, want the create failure")
	}
	if got, want := strings.Join(subcommands(recordedCalls(t, log)), " "), "info image create"; got != want {
		t.Fatalf("docker calls = %q, want %q (no start, nothing to remove)", got, want)
	}
}

func TestCreateAndStartRemovesTheContainerWhenStartFails(t *testing.T) {
	log := useRecordingDocker(t, "start")

	if _, err := CreateAndStart(t.Context()); err == nil {
		t.Fatal("CreateAndStart() error = nil, want the start failure")
	}
	calls := recordedCalls(t, log)
	if got, want := strings.Join(subcommands(calls), " "), "info image create start rm"; got != want {
		t.Fatalf("docker calls = %q, want %q", got, want)
	}
	_, started, _ := strings.Cut(calls[3], " ")
	if want := "rm --force " + started; calls[4] != want {
		t.Errorf("removal call = %q, want it to remove the container that failed to start: %q", calls[4], want)
	}
}
