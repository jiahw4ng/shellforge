// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/lessons"
	"shellforge/internal/storage"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
)

// State contains the grouped state for Shellforge's UI domains.
type State struct {
	// nav tracks the active screen and its current selection.
	nav navigationState
	// lessons holds loaded lesson definitions, progress, and completion data.
	lessons lessonState
	// settings holds state for the Settings and reset-confirmation screens.
	settings settingsState
	// preferences holds learner choices that outlive a visit to Settings.
	preferences preferencesState
	// term owns the active terminal session and its lifecycle state.
	term terminalState
	// viewport stores the latest terminal dimensions reported by Bubble Tea.
	viewport viewportState
	// help controls whether the current screen renders compact or expanded shortcuts.
	help help.Model
}

type navigationState struct {
	// screen is the screen the user is currently viewing.
	screen config.Screen
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
	// guide provides vertical scrolling for the active page's instructions.
	guide viewport.Model
	// revealedHints is the number of active-lesson hints currently visible.
	revealedHints int
	// hintsExpanded controls whether revealed hints are shown in the lesson footer.
	hintsExpanded bool
	// progress is the latest assertion-check state for the active lesson.
	progress progressState
	// completed records lesson IDs whose completion has been confirmed.
	completed map[string]bool
	// progressNotLoaded reports that saved completions could not be read, so
	// completed lessons may be missing their checkmark.
	progressNotLoaded bool
	// progressNotSaved reports that a completion could not be written. It stays
	// set until a successful reset, because earlier completions may be missing too.
	progressNotSaved bool
	// store persists completed lesson IDs when persistence is configured.
	store storage.CompletionStore
}

type progressState struct {
	// results contains the outcomes of the most recent assertion check.
	results []assertion.Result
	// isChecking reports whether an assertion command is currently running.
	isChecking bool
	// hasChecked distinguishes no check yet from a completed check with no results.
	hasChecked bool
	// assertionsExpanded controls whether completed results are shown in the lesson footer.
	assertionsExpanded bool
}

type settingsState struct {
	// message is the latest success or failure text shown in Settings.
	message string
	// failed marks message as an error rather than a success.
	failed bool
	// isResetting prevents repeated confirmation input while reset is running.
	isResetting bool
}

type preferencesState struct {
	// cursorStyleID is the selected ui.CursorStyle ID; empty means the default.
	cursorStyleID string
	// themeID is the selected ui.Theme ID; empty means the default.
	themeID string
	// store persists learner preferences when persistence is configured.
	store storage.SettingsStore
}

// theme returns the selected theme, falling back to the default for an empty
// or unknown ID.
func (p preferencesState) theme() ui.Theme {
	if theme, ok := ui.ThemeByID(p.themeID); ok {
		return theme
	}
	return ui.DefaultTheme
}

// cursorStyle returns the selected cursor style, falling back to the default
// for an empty or unknown ID.
func (p preferencesState) cursorStyle() ui.CursorStyle {
	if style, ok := ui.CursorStyleByID(p.cursorStyleID); ok {
		return style
	}
	return ui.DefaultCursorStyle
}

type terminalState struct {
	// session is the active terminal session, if one has started successfully.
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
