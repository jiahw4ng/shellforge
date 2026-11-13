//go:build !windows

package container

import (
	"testing"
)

// TestListManagedFiltersByShellforgeLabel checks that only labelled containers
// are asked for, stopped ones included, and that Docker's lines are parsed.
func TestListManagedFiltersByShellforgeLabel(t *testing.T) {
	useFakeDocker(t, `case "$1" in
  info) exit 0 ;;
  ps)
    [ "$2" = "--all" ] || exit 2
    [ "$3" = "--filter" ] && [ "$4" = "label=shellforge.managed=true" ] || exit 3
    printf 'abc123\tshellforge-one\trunning\t2026-11-13 10:00:00 +0000 UTC\n'
    ;;
  *) exit 1 ;;
esac
`)

	sandboxes, err := ListManaged(t.Context())
	if err != nil {
		t.Fatalf("ListManaged() error = %v", err)
	}
	if len(sandboxes) != 1 || sandboxes[0].Name != "shellforge-one" || sandboxes[0].State != "running" {
		t.Errorf("ListManaged() = %+v, want one running shellforge-one", sandboxes)
	}
}

func TestListManagedReportsUnreachableDocker(t *testing.T) {
	useFakeDocker(t, `echo "Cannot connect to the Docker daemon" >&2
exit 1
`)

	_, err := ListManaged(t.Context())
	if unavailable := unavailableReason(t, err); unavailable.Reason != ReasonDockerUnreachable {
		t.Errorf("reason = %v, want %v", unavailable.Reason, ReasonDockerUnreachable)
	}
}
