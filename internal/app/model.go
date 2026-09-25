// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/completion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"

	"charm.land/bubbles/v2/spinner"
)

// State contains the grouped state for Shellforge's UI domains.
type State struct {
	// nav tracks the active screen and its current selection.
	nav navigationState
	// lessons holds loaded lesson definitions, progress, and completion data.
	lessons lessonState
	// settings holds state for the Settings and reset-confirmation screens.
	settings settingsState
	// term owns the active terminal session and its lifecycle state.
	term terminalState
	// viewport stores the latest terminal dimensions reported by Bubble Tea.
	viewport viewportState
}

type navigationState struct {
	// screen is the screen the user is currently viewing.
	screen screen
	// selection is the selected item on the current navigable screen.
	selection int
}

type lessonState struct {
	// available is the list of lessons loaded from embedded lesson files.
	available []lessons.Lesson
	// err is the error produced while loading embedded lesson files.
	err error
	// activeIdx is the index of the lesson currently open in lesson mode.
	activeIdx int
	// activePage is the index of the instructional page currently displayed.
	activePage int
	// progress is the latest assertion-check state for the active lesson.
	progress progressState
	// completed records lesson IDs whose completion has been confirmed.
	completed map[string]bool
	// store persists completed lesson IDs when persistence is configured.
	store completion.CompletionStore
}

type progressState struct {
	// results contains the outcomes of the most recent assertion check.
	results []assertion.Result
	// isChecking reports whether an assertion command is currently running.
	isChecking bool
	// hasChecked distinguishes no check yet from a completed check with no results.
	hasChecked bool
}

type settingsState struct {
	// message is the latest success or failure text shown in Settings.
	message string
	// failed marks message as an error rather than a success.
	failed bool
	// isResetting prevents repeated confirmation input while reset is running.
	isResetting bool
}

type terminalState struct {
	// session is the active Bubbleterm session, if one has started successfully.
	session *terminal.TermSession
	// err is the latest terminal-start or unexpected-exit error.
	err error
	// output preserves the final terminal frame after an unexpected exit.
	output string
	// hasRequestedExit records Ctrl+D so its exit is treated as intentional.
	hasRequestedExit bool
	// isStarting reports that a terminal-start command is in progress.
	isStarting bool
	// gen identifies one terminal-start attempt. Asynchronous exit and
	// assertion messages carry it, allowing the app to ignore events left over
	// from a sandbox that was reset or replaced.
	gen uint64
	// spinner animates while a terminal-start command is in progress.
	spinner spinner.Model
}

type viewportState struct {
	// width is the latest reported terminal width in cells.
	width int
	// height is the latest reported terminal height in cells.
	height int
}
