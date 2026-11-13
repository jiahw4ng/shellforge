package container

import (
	"errors"
	"fmt"
	"time"
)

const (
	// stepTimeout bounds each Docker step except the image download.
	stepTimeout = 20 * time.Second
	// pullTimeout is longer because a first-time download crosses the network.
	pullTimeout = 5 * time.Minute
	// inspectTimeout bounds the quick running-state check after a shell exits.
	inspectTimeout = 5 * time.Second

	// imageRepository is where the lesson sandbox image is published (GHCR).
	imageRepository = "ghcr.io/jiahw4ng/shellforge-sandbox"

	// imageVersion is the image tag this build pulls.
	// !!! CAUTION !!!
	// Bump it for EVERY release, even when docker/lesson.Dockerfile did not
	// change. A release tag must be exactly "v" + imageVersion (v0.1.4 for
	// "0.1.4"): publish-image.yml and release.yml read this line, fail on a
	// mismatch, and only a v* tag publishes the image tag. Keeping the two in
	// lockstep is a deliberate trade-off: a version is published once and not
	// reused, so a released binary pulls the exact image it shipped with.
	//
	// It must be plain MAJOR.MINOR.PATCH, with no pre-release suffix:
	// TestImageVersionIsPlainSemver and both release workflows enforce it.
	//
	// When you change it, also update SANDBOX_IMAGE in the Makefile and the image
	// name in the README. The workflows find both constants by name, so keep each
	// on its own line as a plain string.
	imageVersion = "0.1.4"

	// image is the full reference passed to Docker.
	image = imageRepository + ":" + imageVersion

	// ignoreEOFEnvironment is empty-valued on purpose: Bash then uses its default
	// of ignoring 10 consecutive Ctrl+D presses on an empty prompt, so only exit
	// (or an 11th Ctrl+D) ends the learner's shell.
	ignoreEOFEnvironment = "IGNOREEOF="

	// other constants
	dockerExecutable                  = "docker"
	containerNamePrefix               = "shellforge-"
	managedContainerLabel             = "shellforge.managed=true"
	containerSessionLabelPrefix       = "shellforge.session="
	managedListFormat                 = "{{.ID}}\t{{.Names}}\t{{.State}}\t{{.CreatedAt}}"
	managedListFieldCount             = 4
	dockerCreatedAtLayout             = "2006-01-02 15:04:05 -0700"
	containerNetworkMode              = "none"
	containerMemoryLimit              = "256m"
	containerCPULimit                 = "0.5"
	containerPIDLimit                 = "128"
	setupUser                         = "root"
	learnerUser                       = "student"
	learnerWorkspace                  = "/home/student/workspace"
	bashExecutable                    = "/bin/bash"
	learnerBashRC                     = "/etc/shellforge/bashrc"
	interactiveTerminalEnvironment    = "TERM=xterm-256color"
	dockerCLIHooksEnvironment         = "DOCKER_CLI_HOOKS=false"
	dockerCLIHintsEnvironment         = "DOCKER_CLI_HINTS=false"
	logContainerKey                   = "container"
	logErrorKey                       = "error"
	errDockerNotOnPathMsg             = "Docker CLI is not on PATH"
	errDockerUnreachableMsg           = "Docker is not reachable; start it with Linux containers enabled"
	imageBuildCommandFormat           = "docker build --file docker/lesson.Dockerfile --tag %s ."
	errImageMissingMsgFormat          = "could not download the lesson image %s; check your internet connection or build it with `%s`"
	errCannotRemoveLessonContainerMsg = "cannot remove lesson container"
	errCannotRunLessonSetupMsg        = "cannot run lesson setup"
)

// ErrContainerUnavailable is the category every ContainerUnavailableError matches with
// errors.Is. Match on it; do not display it, because the error's own text
// carries the reason and the underlying Docker output.
var ErrContainerUnavailable = errors.New("docker lesson containers are unavailable")

// ImageBuildCommand returns the command that builds the local lesson image.
func ImageBuildCommand() string {
	return fmt.Sprintf(imageBuildCommandFormat, image)
}
