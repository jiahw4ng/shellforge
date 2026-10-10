package terminal

import (
	"errors"
	"fmt"
)

// ErrInvalidStartResult reports an impossible terminal-start result containing
// neither a session nor an error.
var ErrInvalidStartResult = errors.New("terminal startup returned no session and no error")

// StartError identifies the stage that prevented a learner terminal from
// starting while preserving the underlying cause for logs and errors.
type StartError struct {
	Stage string
	Err   error
}

func (e *StartError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap exposes the underlying cause so callers can classify terminal-start
// failures with errors.Is or errors.As.
func (e *StartError) Unwrap() error {
	return e.Err
}
