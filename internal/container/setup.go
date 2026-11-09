package container

import (
	"context"
	"os/exec"
)

// checkPrerequisites confirms that Docker is installed, reachable, and has the
// lesson image, reporting the first prerequisite that is missing.
func checkPrerequisites(ctx context.Context, progress ProgressFunc) error {
	// STEP 1: check docker is installed and available
	progress.Report(StepCheckingDocker)
	if err := checkIsDockerInstalled(); err != nil {
		return err
	}
	if err := checkIsDockerAvailable(ctx); err != nil {
		return err
	}
	return checkIsImageAvailable(ctx, progress)
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
	if err := runDockerCommandWithin(ctx, stepTimeout, "info", "--format", "{{.ServerVersion}}"); err != nil {
		return &ContainerUnavailableError{Reason: ReasonDockerUnreachable, Cause: err}
	}
	return nil
}

// checkIsImageAvailable confirms that the lesson image is on this machine,
// downloading it from the registry when it is not.
func checkIsImageAvailable(ctx context.Context, progress ProgressFunc) error {
	// STEP 2: check lesson image is available
	progress.Report(StepCheckingImage)
	if err := runDockerCommandWithin(ctx, stepTimeout, "image", "inspect", image); err == nil {
		return nil
	}
	// STEP 3: lesson image is missing, pull it from the registry
	progress.Report(StepPullingImage)
	if err := runDockerCommandWithin(ctx, pullTimeout, "pull", "--platform", containerPlatform, image); err != nil {
		return &ContainerUnavailableError{Reason: ReasonImageMissing, Cause: err}
	}
	return nil
}
