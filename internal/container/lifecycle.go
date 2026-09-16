// Package container manages disposable Docker lesson containers.
package container

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os/exec"
	"time"
)

// CreateAndStart creates and starts a uniquely named lesson container.
func CreateAndStart(ctx context.Context) (*Container, error) {
	slog.Debug("checking Docker prerequisites")
	if _, err := exec.LookPath("docker"); err != nil {
		slog.Error("Docker CLI is not on PATH", "error", err)
		return nil, fmt.Errorf("%w: Docker CLI is not on PATH", ErrUnavailable)
	}
	if err := checkDockerAvailable(ctx); err != nil {
		return nil, err
	}
	if err := checkImageAvailable(ctx); err != nil {
		return nil, err
	}

	name, err := createContainerName()
	if err != nil {
		return nil, err
	}
	container := &Container{Name: name}
	slog.Info("creating lesson container", "container", name)

	if err := executeDockerCommand(ctx, getCreateContainerArguments(name)...); err != nil {
		slog.Error("could not create lesson container", "container", name, "error", err)
		return nil, err
	}
	if err := executeDockerCommand(ctx, "start", name); err != nil {
		slog.Error("could not start lesson container", "container", name, "error", err)
		container.Remove(context.Background())
		return nil, err
	}
	slog.Info("lesson container started", "container", name)

	return container, nil
}

// ShellCommand returns the interactive Docker attachment that the PTY runs.
func (c *Container) ShellCommand() *exec.Cmd {
	return exec.Command("docker",
		"exec", "--interactive", "--tty",
		"--user", "student",
		"--workdir", "/home/student/workspace",
		"--env", "TERM=xterm-256color",
		"--env", "PS1=shellforge$ ",
		c.Name,
		"/bin/bash", "--noprofile", "--norc", "-i",
	)
}

// Remove force-removes exactly this session's generated container name.
func (c *Container) Remove(ctx context.Context) {
	c.RemoveOnce.Do(func() {
		slog.Info("removing lesson container", "container", c.Name)
		removeContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := executeDockerCommand(removeContext, "rm", "--force", c.Name); err != nil {
			slog.Error("could not remove lesson container", "container", c.Name, "error", err)
		}
	})
}

// getCreateContainerArguments returns the Docker flags that apply Shellforge's
// isolation limits to one newly created lesson container.
func getCreateContainerArguments(name string) []string {
	return []string{
		"create",
		"--platform", "linux/amd64",
		"--name", name,
		"--label", "shellforge.managed=true",
		"--label", "shellforge.session=" + name,
		"--network", "none",
		"--memory", "256m",
		"--cpus", "0.5",
		"--pids-limit", "128",
		Image,
		"sleep", "infinity",
	}
}

// checkDockerAvailable confirms that the Docker CLI can reach Docker Desktop.
func checkDockerAvailable(ctx context.Context) error {
	if err := executeDockerCommand(ctx, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return fmt.Errorf("%w: Docker Desktop is not reachable; enable WSL integration for this distribution", ErrUnavailable)
	}
	return nil
}

// checkImageAvailable confirms that the locally built lesson image exists.
func checkImageAvailable(ctx context.Context) error {
	if err := executeDockerCommand(ctx, "image", "inspect", Image); err != nil {
		return fmt.Errorf("%w: build the lesson image with `make sandbox-image`", ErrUnavailable)
	}
	return nil
}

// executeDockerCommand runs one Docker CLI command and includes its output in
// any returned error so callers can show a useful diagnostic.
func executeDockerCommand(ctx context.Context, arguments ...string) error {
	command := exec.CommandContext(ctx, "docker", arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		slog.Debug("Docker command failed", "arguments", arguments, "error", err)
		return fmt.Errorf("docker %v: %w: %s", arguments, err, string(output))
	}
	return nil
}

// createContainerName generates an unpredictable name so concurrent lessons do
// not collide and cleanup can target one exact container.
func createContainerName() (string, error) {
	identifier := make([]byte, 8)
	if _, err := rand.Read(identifier); err != nil {
		return "", err
	}
	return "shellforge-" + hex.EncodeToString(identifier), nil
}
