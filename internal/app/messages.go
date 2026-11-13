package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
	"time"
)

type lessonsLoadedMsg struct {
	lessons []lessons.Lesson
	err     error
}

type completionsLoadedMsg struct {
	lessonIDs []string
	err       error
}

type termStartedMsg struct {
	session *terminal.TermSession
	err     error
	gen     uint64
}

// termStartProgressMsg reports that a start attempt has reached a new step.
type termStartProgressMsg struct {
	step container.Step
	gen  uint64
}

// termExitedMsg reports that a session's shell ended. sandboxRunning is true
// only when Docker confirmed the container was still running afterwards,
// meaning the learner left the shell rather than the sandbox failing.
type termExitedMsg struct {
	gen            uint64
	sandboxRunning bool
}

// assertionsCheckedMsg carries the results of an asynchronous lesson check.
type assertionsCheckedMsg struct {
	lessonID string
	results  []assertion.Result
	gen      uint64
}

type completionSavedMsg struct {
	lessonID string
	err      error
}

type completionsResetMsg struct {
	err error
}

type settingsResetMsg struct {
	err error
}

type cursorStyleSavedMsg struct {
	styleID string
	err     error
}

// sandboxesListedMsg carries the result of one Manage sandboxes listing.
type sandboxesListedMsg struct {
	sandboxes []container.ManagedSandbox
	err       error
	listedAt  time.Time
	gen       uint64
}

// sandboxesDeletedMsg reports a deletion of ticked sandboxes: the IDs Docker
// could not remove, or why nothing was removed.
type sandboxesDeletedMsg struct {
	failed []string
	err    error
}

type themeSavedMsg struct {
	themeID string
	err     error
}
