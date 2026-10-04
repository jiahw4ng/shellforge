package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress leaves state unchanged while a reset is running and otherwise
// delegates the keypress to the handler for the active screen.
func (s State) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.settings.isResetting {
		return s, nil
	}
	if s.nav.screen != sandboxTerminalScreen && key.Matches(msg, keys.toggleHelp) {
		s.help.ShowAll = !s.help.ShowAll
		return s, nil
	}

	switch s.nav.screen {
	case lessonTerminalScreen:
		return s.handleLessonKeyPress(msg)
	case sandboxTerminalScreen:
		return s.handleTerminalKeyPress(msg)
	case resetConfirmationScreen:
		return s.handleResetConfirmationKeyPress(msg)
	default:
		return s.handleNavigationKeyPress(msg)
	}
}

// handleNavigationKeyPress updates menu selection or screen state, quits when
// requested, and schedules terminal startup after entering a terminal screen.
func (s State) handleNavigationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// handle ordinary navigation and quit requests, returning early if the app should quit
	if shouldQuit := s.handleNavigation(msg); shouldQuit {
		return s, tea.Quit
	}

	// after navigating, check if the user entered a screen that
	// requires starting a terminal session
	switch s.nav.screen {
	case lessonTerminalScreen:
		return s, s.startTerminalWithActiveLesson()
	case sandboxTerminalScreen:
		return s, s.startTerminal(nil)
	default:
		return s, nil
	}
}

// handleResetConfirmationKeyPress delegates ordinary navigation, reports unavailable
// storage, or marks reset as active and schedules completion deletion.
func (s State) handleResetConfirmationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// if the user did not press Enter on the correct option, handle navigation normally
	if !key.Matches(msg, keys.selectItem) || s.nav.selection != 1 {
		return s.handleNavigationKeyPress(msg)
	}

	// guard againt unavailable storage
	if s.lessons.store == nil {
		s.nav.screen = settingsScreen
		s.nav.selection = 0
		s.settings.message = "Lesson progress could not be reset because storage is unavailable."
		s.settings.failed = true
		return s, nil
	}

	s.settings.isResetting = true
	return s, resetLessonCompletions(s.lessons.store)
}

// handleLessonKeyPress updates lesson progress, sandbox, hints, page, or guide state during a lesson
// forwards all other keys to the embedded terminal.
func (s State) handleLessonKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.checkProgress):
		// check assertions
		if s.term.session == nil || s.lessons.progress.isChecking {
			return s, nil
		}
		s.lessons.progress.isChecking = true
		lesson := s.lessons.available[s.lessons.activeIdx]
		return s, checkAssertions(s.term.session, lesson.ID, lesson.Assertions, s.term.gen)
	case key.Matches(msg, keys.resetSandbox):
		// reset the sandbox
		if s.term.isStarting {
			return s, nil
		}
		s.CloseTerminal()
		return s, s.startTerminalWithActiveLesson()
	case key.Matches(msg, keys.showHint):
		// reveal next hint
		hints := s.lessons.available[s.lessons.activeIdx].Hints
		if s.lessons.revealedHints < len(hints) {
			s.lessons.revealedHints++
			s.prepareLessonGuide()
			s.lessons.guide.GotoBottom()
		}
		return s, nil
	case key.Matches(msg, keys.previousPage):
		// handle previous page navigation within the lesson
		s.handleLessonPreviousPageKeyPress()
		return s, nil
	case key.Matches(msg, keys.nextPage):
		// handle next page navigation within the lesson
		s.handleLessonNextPageKeyPress()
		return s, nil
	case key.Matches(msg, keys.guidePageUp):
		// scroll the lesson guide up
		s.prepareLessonGuide()
		s.lessons.guide.ScrollUp(1)
		return s, nil
	case key.Matches(msg, keys.guidePageDown):
		// scroll the lesson guide down
		s.prepareLessonGuide()
		s.lessons.guide.ScrollDown(1)
		return s, nil
	default:
		// delegate all other keypresses to the embedded terminal
		return s.handleTerminalKeyPress(msg)
	}
}

func (s *State) handleLessonPreviousPageKeyPress() {
	if s.lessons.activePage > 0 {
		s.lessons.activePage--
		s.lessons.guide.GotoTop()

	}
}

func (s *State) handleLessonNextPageKeyPress() {
	pages := s.lessons.available[s.lessons.activeIdx].Pages
	if s.lessons.activePage < len(pages)-1 {
		s.lessons.activePage++
		s.lessons.guide.GotoTop()
	}
}

// handleTerminalKeyPress forwards input to an active session, records a requested
// exit, or clears a terminal failure while returning to its parent screen.
// note: this is used by both the lesson and standalone sandbox terminal.
func (s State) handleTerminalKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.term.session != nil {
		if key.Matches(msg, keys.returnBack) {
			s.term.hasRequestedExit = true
		}
		return s, s.term.session.Update(msg)
	}
	if key.Matches(msg, keys.quit) {
		return s, tea.Quit
	}
	if key.Matches(msg, keys.returnBack) && s.term.err != nil {
		s.returnFromTerminal()
		s.term.err = nil
		s.term.output = ""
	}
	return s, nil
}
