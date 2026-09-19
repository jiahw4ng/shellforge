package terminal

import "fmt"

// StartError identifies the stage that prevented a learner terminal from
// starting while preserving the underlying cause for logs and errors.
type StartError struct {
	Stage string
	Err   error
}

func (e *StartError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

// Unwrap exposes the underlying error to callers that need to classify it.
func (e *StartError) Unwrap() error {
	return e.Err
}
