// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
)

type screen int

const (
	menuScreen screen = iota
	lessonsScreen
	lessonScreen
	featureScreen
	terminalScreen
)

var menuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
}

// State contains Shellforge's navigation, lesson, window, and terminal state.
type State struct {
	CurrentScreen   screen
	SelectedOption  int
	SelectedLesson  int
	ActiveLesson    int
	FeatureReturnTo screen
	Lessons         []lessons.Lesson
	LessonErr       error
	Terminal        *terminal.Session
	Width           int
	Height          int
	TerminalErr     error
}

// New creates the initial main-menu state.
func New() State {
	return State{}
}
