//go:build !windows

package container

import (
	"slices"
	"testing"
)

// fakeManagedDocker lists aaa and bbb as labelled containers and fails to
// remove bbb. Any rm of an unlabelled ID exits 9, so the test would see it.
const fakeManagedDocker = `case "$1" in
  info) exit 0 ;;
  ps)
    [ "$2" = "--all" ] && [ "$3" = "--quiet" ] || exit 2
    [ "$4" = "--filter" ] && [ "$5" = "label=shellforge.managed=true" ] || exit 3
    printf 'aaa\nbbb\n'
    ;;
  rm)
    [ "$2" = "--force" ] || exit 4
    case "$3" in
      aaa) exit 0 ;;
      bbb) echo "removal failed" >&2; exit 1 ;;
      *) exit 9 ;;
    esac
    ;;
  *) exit 1 ;;
esac
`

// TestRemoveManagedRemovesOnlyLabelledSandboxes checks that a labelled sandbox
// is removed, a failed removal is reported, and an ID Docker does not list
// with Shellforge's label is reported without being removed.
func TestRemoveManagedRemovesOnlyLabelledSandboxes(t *testing.T) {
	useFakeDocker(t, fakeManagedDocker)

	failed, err := RemoveManaged(t.Context(), []string{"aaa", "bbb", "zzz"})
	if err != nil {
		t.Fatalf("RemoveManaged() error = %v", err)
	}
	if want := []string{"bbb", "zzz"}; !slices.Equal(failed, want) {
		t.Errorf("RemoveManaged() failed = %v, want %v", failed, want)
	}
}

func TestRemoveManagedReportsUnreachableDocker(t *testing.T) {
	useFakeDocker(t, `echo "Cannot connect to the Docker daemon" >&2
exit 1
`)

	_, err := RemoveManaged(t.Context(), []string{"aaa"})
	if unavailable := unavailableReason(t, err); unavailable.Reason != ReasonDockerUnreachable {
		t.Errorf("reason = %v, want %v", unavailable.Reason, ReasonDockerUnreachable)
	}
}

// TestRemoveManagedRemovesNothingWhenListingFails checks that without a
// labelled listing to check against, no removal is attempted.
func TestRemoveManagedRemovesNothingWhenListingFails(t *testing.T) {
	useFakeDocker(t, `case "$1" in
  info) exit 0 ;;
  ps) exit 1 ;;
  *) exit 9 ;;
esac
`)

	failed, err := RemoveManaged(t.Context(), []string{"aaa"})
	if err == nil {
		t.Fatalf("RemoveManaged() error = nil, failed = %v; want the listing error", failed)
	}
}
