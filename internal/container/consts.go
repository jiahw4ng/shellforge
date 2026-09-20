package container

import "errors"

const (
	image                             = "shellforge-sandbox:0.1.0"
	errDockerNotOnPathMsg             = "Docker CLI is not on PATH"
	errCannotCreateLessonContainerMsg = "cannot create lesson container"
	errCannotStartLessonContainerMsg  = "cannot start lesson container"
	errCannotRemoveLessonContainerMsg = "cannot remove lesson container"
	errCannotRunLessonSetupMsg        = "cannot run lesson setup"
)

var ErrContainerUnavailable = errors.New("docker lesson containers are unavailable")
