package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
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

type termExitedMsg struct {
	gen uint64
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

type themeSavedMsg struct {
	themeID string
	err     error
}
