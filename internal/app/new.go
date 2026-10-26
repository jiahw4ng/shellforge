package app

import (
	"shellforge/internal/storage"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// New creates the initial main-menu state without persistence.
func New() State {
	return NewWithStorage(nil)
}

// NewWithStorage creates the initial state with lesson and settings
// persistence. A nil storage keeps all data in memory only.
func NewWithStorage(store storage.Storage) State {
	return State{
		lessons: lessonState{
			completed: make(map[string]bool),
			store:     store,
			guide:     viewport.New(),
		},
		preferences: preferencesState{store: store},
		term:        terminalState{spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
		help:        newHelpModel(),
	}
}
