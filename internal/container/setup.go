package container

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os/exec"
)

// checkPrerequisites confirms that Docker is installed, reachable, and has the
// lesson image, reporting the first prerequisite that is missing.
func checkPrerequisites(ctx context.Context) error {
	if err := checkIsDockerInstalled(); err != nil {
		return err
	}
	if err := checkIsDockerAvailable(ctx); err != nil {
		return err
	}
	return checkIsImageAvailable(ctx)
}

// checkIsDockerInstalled confirms that the Docker CLI is on PATH.
func checkIsDockerInstalled() error {
	if _, err := exec.LookPath(dockerExecutable); err != nil {
		return &ContainerUnavailableError{Reason: ReasonDockerMissing, Cause: err}
	}
	return nil
}

// checkIsDockerAvailable confirms that the Docker CLI can reach Docker Desktop.
func checkIsDockerAvailable(ctx context.Context) error {
	if err := runDockerCommand(ctx, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return &ContainerUnavailableError{Reason: ReasonDockerUnreachable, Cause: err}
	}
	return nil
}

// checkIsImageAvailable confirms that the lesson image is on this machine,
// downloading it from the registry when it is not.
func checkIsImageAvailable(ctx context.Context) error {
	if err := runDockerCommand(ctx, "image", "inspect", image); err == nil {
		return nil
	}
	if err := runDockerCommand(ctx, "pull", "--platform", containerPlatform, image); err != nil {
		return &ContainerUnavailableError{Reason: ReasonImageMissing, Cause: err}
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
	return containerNamePrefix + hex.EncodeToString(identifier), nil
}
