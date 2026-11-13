//go:build !windows

package container

import (
	"errors"
	"testing"
)

func TestIsRunningReadsDockerState(t *testing.T) {
	for _, test := range []struct {
		state string
		want  bool
	}{
		{state: "true", want: true},
		{state: "false", want: false},
	} {
		t.Run(test.state, func(t *testing.T) {
			useFakeDocker(t, `[ "$1" = "inspect" ] || exit 2
echo "`+test.state+`"
`)

			got, err := (&Container{Name: "shellforge-test"}).IsRunning(t.Context())
			if err != nil || got != test.want {
				t.Errorf("IsRunning() = %v, %v; want %v, nil", got, err, test.want)
			}
		})
	}
}

func TestIsRunningReportsMissingContainer(t *testing.T) {
	useFakeDocker(t, `echo "Error: No such container: shellforge-gone" >&2
exit 1
`)

	running, err := (&Container{Name: "shellforge-gone"}).IsRunning(t.Context())
	var command *DockerCommandError
	if running || !errors.As(err, &command) {
		t.Errorf("IsRunning() = %v, %v; want false and a *DockerCommandError", running, err)
	}
}

func TestIsRunningAsksAboutExactlyThisContainer(t *testing.T) {
	log := useRecordingDocker(t, "")

	if _, err := (&Container{Name: "shellforge-test"}).IsRunning(t.Context()); err != nil {
		t.Fatalf("IsRunning() error = %v", err)
	}
	calls := recordedCalls(t, log)
	want := "inspect --type container --format {{.State.Running}} shellforge-test"
	if len(calls) != 1 || calls[0] != want {
		t.Errorf("docker calls = %q, want [%q]", calls, want)
	}
}
