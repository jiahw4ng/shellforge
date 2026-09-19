// Package app contains Shellforge's Bubble Tea application state machine.
package app

import (
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"

	"charm.land/bubbles/v2/spinner"
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
	CurrentScreen         screen
	SelectedOption        int
	SelectedLesson        int
	ActiveLesson          int
	ActivePage            int
	FeatureReturnTo       screen
	Lessons               []lessons.Lesson
	LessonErr             error
	Terminal              *terminal.TermSession
	Width                 int
	Height                int
	TerminalErr           error
	TerminalOutput        string
	TerminalExitRequested bool
	TerminalStarting      bool
	TerminalSpinner       spinner.Model
}

// New creates the initial main-menu state.
func New() State {
	return State{TerminalSpinner: spinner.New(spinner.WithSpinner(spinner.Dot))}
}
