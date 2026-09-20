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

	// check that the Docker CLI is on PATH and can reach Docker Desktop.
	if _, err := exec.LookPath("docker"); err != nil {
		slog.Error("Docker CLI is not on PATH", "error", err)
		return nil, fmt.Errorf("%w: %s", ErrContainerUnavailable, errDockerNotOnPathMsg)
	}

	// check that Docker Desktop is running and can be reached by the CLI.
	if err := checkIsDockerAvailable(ctx); err != nil {
		return nil, err
	}

	// check that the lesson image is built and available locally.
	if err := checkIsImageAvailable(ctx); err != nil {
		return nil, err
	}

	// create a unique container name
	name, err := createContainerName()
	if err != nil {
		return nil, err
	}
	container := &Container{Name: name}

	slog.Info("creating lesson container", "container", name)

	// create the container
	if err := runDockerCommand(ctx, getCreateContainerArguments(name)...); err != nil {
		slog.Error(errCannotCreateLessonContainerMsg, "container", name, "error", err)
		return nil, err
	}

	// start the container
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
	/*
		--interactive: allocate a PTY for the learner's Bash session
		--tty: allocate a PTY for the learner's Bash session
		--user student: run as the unprivileged student user
		--workdir /home/student/workspace: start in the container-only workspace
		--env TERM=xterm-256color: ensure colorized output in the PTY
		--noprofile --rcfile /etc/shellforge/bashrc -i: use lesson-provided Bashrc
	*/
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
	if err := runDockerCommand(ctx, getSetupCommandArguments(c.Name, setup)...); err != nil {
		slog.Error(errCannotRunLessonSetupMsg, "container", c.Name, "error", err)
		return err
	}
	return nil
}

// Exec runs a non-interactive command in this container and returns its stdout
// and stderr. It is used for assertions, not for the learner's PTY session.
func (c *Container) Exec(ctx context.Context, user string, workingDir string, command ...string) (string, error) {
	arguments := getExecCommandArguments(c.Name, user, workingDir, command)
	output, err := runDockerCommandWithOutput(ctx, arguments...)
	if err != nil {
		return "", err
	}
	return output, nil
}
