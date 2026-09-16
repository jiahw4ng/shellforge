package tui

import (
	"shellforge/internal/container"
	"shellforge/internal/lessons"

	bubbleterm "github.com/taigrr/bubbleterm"
)

// State is Shellforge's initial navigation state
type State struct {
	TermState
	CurrentScreen   currentScreen
	SelectedOption  int
	SelectedLesson  int
	ActiveLesson    int
	FeatureReturnTo currentScreen
	Lessons         []lessons.Lesson
	LessonContainer *container.Container
	LessonErr       error
}

type TermState struct {
	Terminal   *bubbleterm.Model
	TermWidth  int
	TermHeight int
	TermExit   <-chan struct{}
	TermErr    error
}

// lessonsLoadedMsg carries the result of parsing the embedded lesson files.
type lessonsLoadedMsg struct {
	lessons []lessons.Lesson
	err     error
}
