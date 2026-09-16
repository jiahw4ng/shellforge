package app

import (
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
)

type LessonsLoadedMsg struct {
	Lessons []lessons.Lesson
	Err     error
}

type TerminalStartedMsg struct {
	Session *terminal.Session
	Err     error
}

type TerminalExitedMsg struct{}
