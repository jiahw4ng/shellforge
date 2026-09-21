package assertion

import "context"

// Executor runs a command in the active lesson container. Container satisfies
// this interface, while tests can use a small fake without requiring Docker.
type Executor interface {
	Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error)
}
