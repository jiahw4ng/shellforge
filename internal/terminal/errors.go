package terminal

import (
	"errors"
	"fmt"
	"shellforge/internal/container"
	"strings"
	"unicode/utf8"
)

const (
	dockerMissingSummary     = "Shellforge couldn't find the Docker command."
	dockerMissingRemedy      = "Install Docker and make sure `docker` is on your PATH."
	dockerUnreachableSummary = "Docker isn't reachable."
	dockerUnreachableRemedy  = "Start Docker with Linux containers enabled."
	imageMissingSummary      = "Shellforge couldn't download the lesson sandbox image."
	imageMissingRemedyFormat = "Check your internet connection and try again. To build the image yourself instead, run:\n\n  %s"
	dockerSaidFormat         = "Docker said: %s"

	// maxDetailRunes keeps one line of Docker's output from filling the pane.
	maxDetailRunes = 200
)

// TerminalStartError identifies the stage that prevented a learner terminal from
// starting while preserving the underlying cause for logs and errors.
type TerminalStartError struct {
	Stage string
	Err   error
}

func (e *TerminalStartError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap exposes the underlying error to callers that need to classify it.
func (e *TerminalStartError) Unwrap() error {
	return e.Err
}

// FailureMessage explains a startup failure in terms a learner can act on. It
// deliberately leaves out internal stage names and error categories, which stay
// in the log.
type FailureMessage struct {
	// Summary says what went wrong; Remedy says what to do about it.
	Summary string
	Remedy  string
	// Detail quotes the first line of Docker's own output ("Docker said: ..."),
	// and is empty when Docker printed nothing.
	Detail string
}

// Describe returns the learner-facing message for a recognised startup
// failure. ok is false for any other error, so callers can show its text as is.
func Describe(err error) (message FailureMessage, ok bool) {
	var unavailable *container.ContainerUnavailableError
	// errors.As walks err's Unwrap chain (TerminalStartError, then whatever it
	// wraps) and stops at the first error of type *ContainerUnavailableError,
	// storing it in unavailable. It returns false, leaving unavailable nil, when
	// no error in the chain has that type, so other failures are not described.
	if !errors.As(err, &unavailable) {
		return FailureMessage{}, false
	}

	switch unavailable.Reason {
	case container.ReasonDockerMissing:
		message = FailureMessage{Summary: dockerMissingSummary, Remedy: dockerMissingRemedy}
	case container.ReasonDockerUnreachable:
		message = FailureMessage{Summary: dockerUnreachableSummary, Remedy: dockerUnreachableRemedy}
	case container.ReasonImageMissing:
		message = FailureMessage{
			Summary: imageMissingSummary,
			Remedy:  fmt.Sprintf(imageMissingRemedyFormat, container.ImageBuildCommand()),
		}
	default:
		return FailureMessage{}, false
	}
	if line := dockerOutputLine(err); line != "" {
		message.Detail = fmt.Sprintf(dockerSaidFormat, line)
	}
	return message, true
}

// dockerOutputLine returns the first non-empty line of the failed Docker
// command's output, shortened to maxDetailRunes.
func dockerOutputLine(err error) string {
	var command *container.DockerCommandError
	if !errors.As(err, &command) {
		return ""
	}
	for _, line := range strings.Split(command.Output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if utf8.RuneCountInString(line) > maxDetailRunes {
			line = string([]rune(line)[:maxDetailRunes]) + "…"
		}
		return line
	}
	return ""
}
