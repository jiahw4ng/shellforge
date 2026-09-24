// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/completion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"

	"charm.land/bubbles/v2/spinner"
)

type screen int

// State contains the grouped state for Shellforge's UI domains.
type State struct {
	Nav      navigationState
	Lessons  lessonState
	Settings settingsState
	Term     terminalState
	Viewport viewportState
}

type navigationState struct {
	// what screen the user is currently viewing
	Screen screen
	// which item is selected on the current navigable screen
	Selection int
}

type lessonState struct {
	// list of available lessons
	Available []lessons.Lesson
	Error     error
	// current active lesson
	ActiveIndex int
	// current active page
	ActivePage int
	Progress   progressState
	Completed  map[string]bool
	Store      completion.CompletionStore
}

type progressState struct {
	// results of the most recent assertion checks
	Results    []assertion.Result
	IsChecking bool
	HasChecked bool
}

type settingsState struct {
	Message     string
	Failed      bool
	IsResetting bool
}

type terminalState struct {
	// the active terminal session, if any
	Session          *terminal.TermSession
	Error            error
	Output           string
	hasRequestedExit bool
	IsStarting       bool
	generation       uint64
	Spinner          spinner.Model
}

type viewportState struct {
	Width  int
	Height int
}
