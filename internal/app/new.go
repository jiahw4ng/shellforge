package app

import (
	"shellforge/internal/completion"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// New creates the initial main-menu state.
func New() State {
	return NewWithStores(nil, nil)
}

// NewWithCompletionStore creates the initial state with lesson persistence.
func NewWithCompletionStore(store completion.CompletionStore) State {
	return NewWithStores(store, nil)
}

// NewWithStores creates the initial state with lesson and settings persistence.
// Either store may be nil, leaving that data in memory only.
func NewWithStores(store completion.CompletionStore, settingsStore completion.SettingsStore) State {
	return State{
		lessons: lessonState{
			completed: make(map[string]bool),
			store:     store,
			guide:     viewport.New(),
		},
		preferences: preferencesState{store: settingsStore},
		term:        terminalState{spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
		help:        newHelpModel(),
	}
}
