// Package container manages disposable Docker lesson containers.
package container

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const Image = "shellforge-sandbox:0.1.0"

var ErrUnavailable = errors.New("docker lesson containers are unavailable")

// LessonContainer is one disposable, isolated lesson environment.
type LessonContainer struct {
	name       string
	removeOnce sync.Once
}

// CreateAndStart creates and starts a uniquely named lesson container.
func CreateAndStart(ctx context.Context) (*LessonContainer, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("%w: Docker CLI is not on PATH", ErrUnavailable)
	}
	if err := dockerAvailable(ctx); err != nil {
		return nil, err
	}
	if err := imageAvailable(ctx); err != nil {
		return nil, err
	}

	name, err := containerName()
	if err != nil {
		return nil, err
	}
	container := &LessonContainer{name: name}

	if err := runDocker(ctx, createArguments(name)...); err != nil {
		return nil, err
	}
	if err := runDocker(ctx, "start", name); err != nil {
		container.Remove(context.Background())
		return nil, err
	}

	return container, nil
}

// ShellCommand returns the interactive Docker attachment that the PTY runs.
func (c *LessonContainer) ShellCommand() *exec.Cmd {
	return exec.Command("docker",
		"exec", "--interactive", "--tty",
		"--user", "student",
		"--workdir", "/home/student/workspace",
		"--env", "TERM=xterm-256color",
		"--env", "PS1=shellforge$ ",
		c.name,
		"/bin/bash", "--noprofile", "--rcfile", "/opt/shellforge/interactive.bashrc", "-i",
	)
}

// Remove force-removes exactly this session's generated container name.
func (c *LessonContainer) Remove(ctx context.Context) {
	c.removeOnce.Do(func() {
		removeContext, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = runDocker(removeContext, "rm", "--force", c.name)
	})
}

func createArguments(name string) []string {
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

func dockerAvailable(ctx context.Context) error {
	if err := runDocker(ctx, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return fmt.Errorf("%w: Docker Desktop is not reachable; enable WSL integration for this distribution", ErrUnavailable)
	}
	return nil
}

func imageAvailable(ctx context.Context) error {
	if err := runDocker(ctx, "image", "inspect", Image); err != nil {
		return fmt.Errorf("%w: build the lesson image with `make sandbox-image`", ErrUnavailable)
	}
	return nil
}

func runDocker(ctx context.Context, arguments ...string) error {
	command := exec.CommandContext(ctx, "docker", arguments...)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("docker %v: %w: %s", arguments, err, string(output))
	}
	return nil
}

func containerName() (string, error) {
	identifier := make([]byte, 8)
	if _, err := rand.Read(identifier); err != nil {
		return "", err
	}
	return "shellforge-" + hex.EncodeToString(identifier), nil
}
