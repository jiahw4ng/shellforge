package container

import "errors"

const (
	image                             = "shellforge-sandbox:0.1.0"
	dockerExecutable                  = "docker"
	containerNamePrefix               = "shellforge-"
	containerPlatform                 = "linux/amd64"
	managedContainerLabel             = "shellforge.managed=true"
	containerSessionLabelPrefix       = "shellforge.session="
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
	errImageMissingMsgFormat          = "build the lesson image with `docker build --file docker/lesson.Dockerfile --platform linux/amd64 --tag %s .`"
	errCannotRemoveLessonContainerMsg = "cannot remove lesson container"
	errCannotRunLessonSetupMsg        = "cannot run lesson setup"
)

// ErrContainerUnavailable is the category every UnavailableError matches with
// errors.Is. Match on it; do not display it, because the error's own text
// carries the reason and the underlying Docker output.
var ErrContainerUnavailable = errors.New("docker lesson containers are unavailable")
