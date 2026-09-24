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
	nav      navigationState
	lessons  lessonState
	settings settingsState
	term     terminalState
	viewport viewportState
}

type navigationState struct {
	// what screen the user is currently viewing
	screen screen
	// which item is selected on the current navigable screen
	selection int
}

type lessonState struct {
	// list of available lessons
	available []lessons.Lesson
	err       error
	// current active lesson
	activeIdx int
	// current active page
	activePage int
	progress   progressState
	completed  map[string]bool
	store      completion.CompletionStore
}

type progressState struct {
	// results of the most recent assertion checks
	results    []assertion.Result
	isChecking bool
	hasChecked bool
}

type settingsState struct {
	message     string
	failed      bool
	isResetting bool
}

type terminalState struct {
	// the active terminal session, if any
	session          *terminal.TermSession
	err              error
	output           string
	hasRequestedExit bool
	isStarting       bool
	generation       uint64
	spinner          spinner.Model
}

type viewportState struct {
	width  int
	height int
}
