package app

import (
	"shellforge/internal/assertion"
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
