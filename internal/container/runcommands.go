package container

import (
	"context"
	"log/slog"
	"os/exec"
	"time"
)

// runDockerCommand runs one Docker CLI command and includes its output in
// any returned error so callers can show a useful diagnostic.
func runDockerCommand(ctx context.Context, arguments ...string) error {
	_, err := runDockerCommandWithOutput(ctx, arguments...)
	return err
}

// runDockerCommandWithin runs one Docker CLI command, giving up after timeout.
func runDockerCommandWithin(ctx context.Context, timeout time.Duration, arguments ...string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return runDockerCommand(ctx, arguments...)
}

// runDockerCommandWithOutput runs one Docker CLI command and returns its
// combined output. Errors retain that output for useful diagnostics.
func runDockerCommandWithOutput(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, dockerExecutable, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		slog.Debug("Docker command failed", "arguments", arguments, logErrorKey, err)
		return "", &DockerCommandError{Arguments: arguments, Err: err, Output: string(output)}
	}
	return string(output), nil
}
