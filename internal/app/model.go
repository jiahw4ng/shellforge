// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"

	"charm.land/bubbles/v2/spinner"
)

type screen int

const (
	menuScreen screen = iota
	lessonsScreen
	lessonScreen
	featureScreen
	terminalScreen
)

var menuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
	"Exit",
}

// State contains the grouped state for Shellforge's UI domains.
type State struct {
	Nav      navigationState
	Lessons  lessonState
	Term     terminalState
	Viewport viewportState
}

type navigationState struct {
	// what screen the user is currently viewing
	Screen screen
	// which item is selected on the current navigable screen
	Selection       int
	FeatureReturnTo screen
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
}

type progressState struct {
	// results of the most recent assertion checks
	Results    []assertion.Result
	isChecking bool
	hasChecked bool
}

type terminalState struct {
	// the active terminal session, if any
	Session          *terminal.TermSession
	Error            error
	Output           string
	hasRequestedExit bool
	isStarting       bool
	Spinner          spinner.Model
}

type viewportState struct {
	Width  int
	Height int
}

// New creates the initial main-menu state.
func New() State {
	return State{Term: terminalState{Spinner: spinner.New(spinner.WithSpinner(spinner.Dot))}}
}
