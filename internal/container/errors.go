package container

import "fmt"

// Reason identifies which prerequisite stopped Docker lesson containers from
// being used.
type Reason int

const (
	// ReasonDockerMissing means the Docker CLI is not on PATH.
	ReasonDockerMissing Reason = iota + 1
	// ReasonDockerUnreachable means the CLI could not reach a Docker engine.
	ReasonDockerUnreachable
	// ReasonImageMissing means the locally built lesson image does not exist.
	ReasonImageMissing
)

func (r Reason) String() string {
	switch r {
	case ReasonDockerMissing:
		return errDockerNotOnPathMsg
	case ReasonDockerUnreachable:
		return errDockerUnreachableMsg
	case ReasonImageMissing:
		return fmt.Sprintf(errImageMissingMsgFormat, ImageBuildCommand())
	default:
		return fmt.Sprintf("unknown reason %d", int(r))
	}
}

// ContainerUnavailableError reports that lesson containers cannot be created because a
// prerequisite is missing. Cause keeps the original error, including Docker's
// own output, so it reaches the log instead of being replaced by a guess.
type ContainerUnavailableError struct {
	Reason Reason
	Cause  error
}

func (e *ContainerUnavailableError) Error() string {
	return fmt.Sprintf("%s: %v", e.Reason, e.Cause)
}

// Unwrap exposes the underlying cause to errors.Is and errors.As.
func (e *ContainerUnavailableError) Unwrap() error {
	return e.Cause
}

// Is makes every ContainerUnavailableError match ErrContainerUnavailable, so callers can
// test for the whole category without knowing the reason.
func (e *ContainerUnavailableError) Is(target error) bool {
	return target == ErrContainerUnavailable
}

// DockerCommandError is a failed Docker CLI invocation. It keeps Docker's own output
// apart from the wrapped error, so callers can show that message without
// parsing the error text.
type DockerCommandError struct {
	Arguments []string
	Err       error
	Output    string
}

func (e *DockerCommandError) Error() string {
	return fmt.Sprintf("%s %v: %v: %s", dockerExecutable, e.Arguments, e.Err, e.Output)
}

// Unwrap exposes the process error, such as an *exec.ExitError.
func (e *DockerCommandError) Unwrap() error {
	return e.Err
}
