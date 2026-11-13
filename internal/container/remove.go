package container

import (
	"context"
	"log/slog"
	"strings"

	"github.com/samber/lo"
)

// RemoveManaged force-removes the given managed sandboxes, which may be
// running or stopped, and returns the IDs it could not remove. An ID is only
// removed when Docker currently lists it with Shellforge's managed label, so a
// container Shellforge did not create is never touched; any other ID, such as
// one already gone, is reported as failed.
//
// Docker being missing or unreachable, or the labelled listing failing, is
// returned as an error and nothing is removed. Docker being missing or
// unreachable is a *ContainerUnavailableError, as it is for ListManaged.
func RemoveManaged(ctx context.Context, ids []string) (failed []string, err error) {
	if err := checkIsDockerInstalled(); err != nil {
		return nil, err
	}
	if err := checkIsDockerAvailable(ctx); err != nil {
		return nil, err
	}
	managed, err := listManagedIDs(ctx)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		if !managed[id] {
			slog.Warn("refusing to remove a container that is not a Shellforge sandbox", logContainerKey, id)
			failed = append(failed, id)
			continue
		}
		slog.Info("removing managed sandbox", logContainerKey, id)
		if err := runDockerCommandWithin(ctx, stepTimeout, getRemoveContainerArguments(id)...); err != nil {
			slog.Error(errCannotRemoveLessonContainerMsg, logContainerKey, id, logErrorKey, err)
			failed = append(failed, id)
		}
	}
	return failed, nil
}

// listManagedIDs returns the short IDs of every container carrying
// Shellforge's managed label, running or stopped.
func listManagedIDs(ctx context.Context) (map[string]bool, error) {
	listContext, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	output, err := runDockerCommandWithOutput(listContext, getListManagedIDArguments()...)
	if err != nil {
		return nil, err
	}
	ids := lo.Compact(strings.Fields(output))
	return lo.SliceToMap(ids, func(id string) (string, bool) { return id, true }), nil
}
