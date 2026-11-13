package container

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ManagedSandbox is one Docker container that Shellforge created, found by its
// shellforge.managed label. It may belong to this window, another Shellforge
// window, or one that exited without cleaning up.
type ManagedSandbox struct {
	// ID is Docker's short container ID.
	ID string
	// Name is the generated container name, such as shellforge-3fa2c1d9e8b7a6f5.
	Name string
	// State is Docker's state word, such as running or exited.
	State string
	// CreatedAt is when Docker created the container.
	CreatedAt time.Time
}

// ListManaged returns every container carrying Shellforge's managed label,
// running or stopped, newest first. Docker being missing or unreachable is
// reported as a *ContainerUnavailableError, as it is for CreateAndStart.
func ListManaged(ctx context.Context) ([]ManagedSandbox, error) {
	if err := checkIsDockerInstalled(); err != nil {
		return nil, err
	}
	if err := checkIsDockerAvailable(ctx); err != nil {
		return nil, err
	}

	listContext, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	output, err := runDockerCommandWithOutput(listContext, getListManagedArguments()...)
	if err != nil {
		return nil, err
	}
	return parseManagedSandboxes(output)
}

// parseManagedSandboxes reads the tab-separated lines produced by
// managedListFormat and orders the sandboxes newest first.
func parseManagedSandboxes(output string) ([]ManagedSandbox, error) {
	var sandboxes []ManagedSandbox
	for line := range strings.Lines(output) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != managedListFieldCount {
			return nil, fmt.Errorf("unexpected docker ps line %q", line)
		}
		createdAt, err := parseCreatedAt(fields[3])
		if err != nil {
			return nil, fmt.Errorf("parse creation time of %s: %w", fields[1], err)
		}
		sandboxes = append(sandboxes, ManagedSandbox{
			ID:        fields[0],
			Name:      fields[1],
			State:     fields[2],
			CreatedAt: createdAt,
		})
	}
	slices.SortStableFunc(sandboxes, func(a, b ManagedSandbox) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})
	return sandboxes, nil
}

// parseCreatedAt reads Docker's CreatedAt, such as "2026-11-13 10:00:00 +0900 KST".
// The trailing zone name is ignored because some zones print a numeric one
// (such as "+09") that time.Parse cannot read; the offset alone fixes the instant.
func parseCreatedAt(value string) (time.Time, error) {
	fields := strings.Fields(value)
	if len(fields) < 3 {
		return time.Time{}, fmt.Errorf("unexpected creation time %q", value)
	}
	return time.Parse(dockerCreatedAtLayout, strings.Join(fields[:3], " "))
}
