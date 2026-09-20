package container

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

// runDockerCommand runs one Docker CLI command and includes its output in
// any returned error so callers can show a useful diagnostic.
func runDockerCommand(ctx context.Context, arguments ...string) error {
	_, err := runDockerCommandWithOutput(ctx, arguments...)
	return err
}

// runDockerCommandWithOutput runs one Docker CLI command and returns its
// combined output. Errors retain that output for useful diagnostics.
func runDockerCommandWithOutput(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, "docker", arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		slog.Debug("Docker command failed", "arguments", arguments, "error", err)
		return "", fmt.Errorf("docker %v: %w: %s", arguments, err, string(output))
	}
	return string(output), nil
}
