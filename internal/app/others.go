package app

import (
	"shellforge/internal/lessons"
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// startTerminalWithActiveLesson creates a fresh sandbox for the lesson currently
// displayed beside the terminal.
func (s *State) startTerminalWithActiveLesson() tea.Cmd {
	lesson := &s.lessons.available[s.lessons.activeIdx]
	return s.startTerminal(lesson)
}

// startTerminal clears session-only state and starts a new terminal attempt.
// Every attempt receives a generation so delayed events from an older session
// cannot alter this one.
func (s *State) startTerminal(lesson *lessons.Lesson) tea.Cmd {
	s.term.gen++
	s.term.output = ""
	s.term.err = nil
	s.term.hasRequestedExit = false
	s.term.isStarting = true
	s.lessons.progress = progressState{}
	s.term.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	width, height := s.terminalDimensions()
	return tea.Batch(startTerminal(width, height, lesson, s.term.gen), s.term.spinner.Tick)
}

// returnFromTerminal returns to the screen that launched the terminal.
func (s *State) returnFromTerminal() {
	s.term.hasRequestedExit = false
	if s.nav.screen == lessonTerminalScreen {
		s.nav.screen = lessonsScreen
		return
	}
	s.nav.screen = menuScreen
}

func (s State) usesTerminal() bool {
	return s.nav.screen == sandboxTerminalScreen || s.nav.screen == lessonTerminalScreen
}

func (s State) resizeTerminal() tea.Cmd {
	width, height := s.terminalDimensions()
	return s.term.session.Resize(width, height)
}

func (s State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	if s.nav.screen == lessonTerminalScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return ui.DimensionWithFallback(width, 80), screens.TerminalContentHeight(height)
}

func (s *State) CloseTerminal() {
	if s.term.session != nil {
		s.term.session.Close()
	}
	s.term.session = nil
}
