package terminal

import "fmt"

// TerminalStartError identifies the stage that prevented a learner terminal from
// starting while preserving the underlying cause for logs and errors.
type TerminalStartError struct {
	Stage string
	Err   error
}

func (e *TerminalStartError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap exposes the underlying cause so callers can classify terminal-start
// failures with errors.Is or errors.As.
func (e *TerminalStartError) Unwrap() error {
	return e.Err
}
