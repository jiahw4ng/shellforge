package app

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// New creates the initial main-menu state.
func New() State {
	return NewWithCompletionStore(nil)
}

// NewWithCompletionStore creates the initial state with lesson persistence.
func NewWithCompletionStore(store CompletionStore) State {
	return State{
		lessons: lessonState{
			completed: make(map[string]bool),
			store:     store,
			guide:     viewport.New(),
		},
		term: terminalState{spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
		help: newHelpModel(),
	}
}
