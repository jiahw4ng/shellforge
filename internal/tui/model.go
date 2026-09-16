package tui

import (
	"shellforge/internal/container"

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
	LessonContainer *container.LessonContainer
}

type TermState struct {
	Terminal   *bubbleterm.Model
	TermWidth  int
	TermHeight int
	TermExit   <-chan struct{}
	TermErr    error
}
