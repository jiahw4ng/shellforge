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
	logContainerKey                   = "container"
	logErrorKey                       = "error"
	errDockerNotOnPathMsg             = "Docker CLI is not on PATH"
	errCannotCreateLessonContainerMsg = "cannot create lesson container"
	errCannotStartLessonContainerMsg  = "cannot start lesson container"
	errCannotRemoveLessonContainerMsg = "cannot remove lesson container"
	errCannotRunLessonSetupMsg        = "cannot run lesson setup"
)

var ErrContainerUnavailable = errors.New("docker lesson containers are unavailable")
