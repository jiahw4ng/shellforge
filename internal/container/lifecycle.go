// Package container manages disposable Docker lesson containers.
package container

import (
	"context"
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
		return nil, fmt.Errorf("%w: %s", ErrContainerUnavailable, errDockerNotOnPathMsg)
	}

	if err := checkIsDockerAvailable(ctx); err != nil {
		return nil, err
	}

	if err := checkIsImageAvailable(ctx); err != nil {
		return nil, err
	}

	name, err := createContainerName()
	if err != nil {
		return nil, err
	}
	container := &Container{Name: name}

	slog.Info("creating lesson container", "container", name)

	if err := runDockerCommand(ctx, createContainerArguments(name)...); err != nil {
		slog.Error(errCannotCreateLessonContainerMsg, "container", name, "error", err)
		return nil, err
	}

	if err := runDockerCommand(ctx, "start", name); err != nil {
		slog.Error(errCannotStartLessonContainerMsg, "container", name, "error", err)
		container.Remove(context.Background())
		return nil, err
	}
	slog.Info("lesson container started", "container", name)

	return container, nil
}

// ShellCommand returns the interactive Docker attachment that the PTY runs.
func (c *Container) ShellCommand() *exec.Cmd {
	// Ignore user profiles and load only the image-provided shell configuration
	// so lesson tracking hooks are deterministic.
	return exec.Command("docker",
		"exec", "--interactive", "--tty",
		"--user", "student",
		"--workdir", "/home/student/workspace",
		"--env", "TERM=xterm-256color",
		c.Name,
		"/bin/bash", "--noprofile", "--rcfile", "/etc/shellforge/bashrc", "-i",
	)
}

// Remove force-removes exactly this session's generated container name.
func (c *Container) Remove(ctx context.Context) {
	c.RemoveOnce.Do(func() {
		slog.Info("removing lesson container", "container", c.Name)
		removeContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := runDockerCommand(removeContext, "rm", "--force", c.Name); err != nil {
			slog.Error(errCannotRemoveLessonContainerMsg, "container", c.Name, "error", err)
		}
	})
}

// RunSetupLesson configures the fresh container from trusted lesson-authored
// Bash before the learner attaches as the unprivileged student user.
func (c *Container) RunSetupLesson(ctx context.Context, setup string) error {
	if setup == "" {
		return nil
	}
	slog.Info("running lesson setup", "container", c.Name)
	if err := runDockerCommand(ctx, setupCommandArguments(c.Name, setup)...); err != nil {
		slog.Error(errCannotRunLessonSetupMsg, "container", c.Name, "error", err)
		return err
	}
	return nil
}

// Exec runs a non-interactive command in this container and returns its stdout
// and stderr. It is used for assertions, not for the learner's PTY session.
func (c *Container) Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error) {
	arguments := execCommandArguments(c.Name, user, workingDir, command)
	output, err := runDockerCommandWithOutput(ctx, arguments...)
	if err != nil {
		return "", err
	}
	return output, nil
}
