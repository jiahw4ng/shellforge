package app

import (
	"shellforge/internal/completion"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// New creates the initial main-menu state.
func New() State {
	return NewWithCompletionStore(nil)
}

// NewWithCompletionStore creates the initial state with lesson persistence.
func NewWithCompletionStore(store completion.CompletionStore) State {
	guide := viewport.New()
	guide.MouseWheelEnabled = false
	guide.KeyMap.PageUp = keys.guidePageUp
	guide.KeyMap.PageDown = keys.guidePageDown
	guide.KeyMap.HalfPageUp.SetEnabled(false)
	guide.KeyMap.HalfPageDown.SetEnabled(false)
	guide.KeyMap.Up.SetEnabled(false)
	guide.KeyMap.Down.SetEnabled(false)
	guide.KeyMap.Left.SetEnabled(false)
	guide.KeyMap.Right.SetEnabled(false)

	return State{
		lessons: lessonState{
			completed: make(map[string]bool),
			store:     store,
			guide:     guide,
		},
		term: terminalState{spinner: spinner.New(spinner.WithSpinner(spinner.Dot))},
	}
}
