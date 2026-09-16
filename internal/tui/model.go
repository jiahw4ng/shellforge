package tui

import (
	"shellforge/internal/container"

	bubbleterm "github.com/taigrr/bubbleterm"
)

// Model is Shellforge's initial navigation state
type Model struct {
	currentScreen    currentScreen
	selectedOption   int
	selectedLesson   int
	activeLesson     int
	featureReturnTo  currentScreen
	termWidth        int
	termHeight       int
	terminal         *bubbleterm.Model
	lessonContainer  *container.LessonContainer
	terminalExit     <-chan struct{}
	terminalStartErr error
}
