package container

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// checkIsDockerAvailable confirms that the Docker CLI can reach Docker Desktop.
func checkIsDockerAvailable(ctx context.Context) error {
	if err := runDockerCommand(ctx, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return fmt.Errorf("%w: Docker Desktop is not reachable; start it with Linux containers enabled", ErrContainerUnavailable)
	}
	return nil
}

// checkIsImageAvailable confirms that the locally built lesson image exists.
func checkIsImageAvailable(ctx context.Context) error {
	if err := runDockerCommand(ctx, "image", "inspect", image); err != nil {
		return fmt.Errorf("%w: build the lesson image with `docker build --file docker/lesson.Dockerfile --platform linux/amd64 --tag %s .`", ErrContainerUnavailable, image)
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
