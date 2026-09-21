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
	Available   []lessons.Lesson
	LoadErr     error
	ActiveIndex int
	ActivePage  int
	Progress    progressState
}

type progressState struct {
	Results  []assertion.Result
	Checking bool
	Checked  bool
}

type terminalState struct {
	Session       *terminal.TermSession
	Error         error
	Output        string
	ExitRequested bool
	Starting      bool
	Spinner       spinner.Model
}

type viewportState struct {
	Width  int
	Height int
}

// New creates the initial main-menu state.
func New() State {
	return State{Term: terminalState{Spinner: spinner.New(spinner.WithSpinner(spinner.Dot))}}
}
