package app

import (
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
)

type LessonsLoadedMsg struct {
	Lessons []lessons.Lesson
	Err     error
}

type CompletionsLoadedMsg struct {
	LessonIDs []string
	Err       error
}

type TerminalStartedMsg struct {
	Session *terminal.TermSession
	Err     error
}

type TerminalExitedMsg struct{}

// AssertionsCheckedMsg carries the results of an asynchronous lesson check.
type AssertionsCheckedMsg struct {
	LessonID string
	Results  []assertion.Result
}

type CompletionSavedMsg struct {
	LessonID string
	Err      error
}

type CompletionsResetMsg struct {
	Err error
}
