package app

import (
	"shellforge/internal/completion"

	"charm.land/bubbles/v2/spinner"
)

// New creates the initial main-menu state.
func New() State {
	return NewWithCompletionStore(nil)
}

// NewWithCompletionStore creates the initial state with lesson persistence.
func NewWithCompletionStore(store completion.CompletionStore) State {
	return State{
		Lessons: lessonState{
			Completed: make(map[string]bool),
			Store:     store,
		},
		Term: terminalState{Spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
	}
}
