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

	switch s.nav.screen {
	case lessonScreen:
		return s.handleLessonKey(msg)
	case terminalScreen:
		return s.handleTerminalKey(msg)
	case resetConfirmationScreen:
		return s.handleResetConfirmationKey(msg)
	default:
		return s.handleNavigationKeyPress(msg)
	}
}

// handleNavigationKeyPress updates menu selection or screen state, quits when
// requested, and schedules terminal startup after entering a terminal screen.
func (s State) handleNavigationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.handleNavigationKey(msg) {
		return s, tea.Quit
	}

	switch s.nav.screen {
	case lessonScreen:
		return s, s.startActiveLessonTerminal()
	case terminalScreen:
		return s, s.startTerminal(nil)
	default:
		return s, nil
	}
}

// handleResetConfirmationKey delegates ordinary navigation, reports unavailable
// storage, or marks reset as active and schedules completion deletion.
func (s State) handleResetConfirmationKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if !key.Matches(msg, keys.selectItem) || s.nav.selection != 1 {
		return s.handleNavigationKeyPress(msg)
	}

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

// handleLessonKey updates lesson progress, sandbox, hints, page, or guide state
// for reserved shortcuts and forwards all other keys to the embedded terminal.
func (s State) handleLessonKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.checkProgress):
		if s.term.session == nil || s.lessons.progress.isChecking {
			return s, nil
		}
		s.lessons.progress.isChecking = true
		lesson := s.lessons.available[s.lessons.activeIdx]
		return s, checkAssertions(s.term.session, lesson.ID, lesson.Assertions, s.term.gen)
	case key.Matches(msg, keys.resetSandbox):
		if s.term.isStarting {
			return s, nil
		}
		s.closeTerminal()
		return s, s.startActiveLessonTerminal()
	case key.Matches(msg, keys.showHint):
		hints := s.lessons.available[s.lessons.activeIdx].Hints
		if s.lessons.revealedHints < len(hints) {
			s.lessons.revealedHints++
			s.prepareLessonGuide()
			s.lessons.guide.GotoBottom()
		}
		return s, nil
	case s.handleLessonPageKey(msg):
		return s, nil
	case key.Matches(msg, keys.guidePageUp):
		s.prepareLessonGuide()
		s.lessons.guide.ScrollUp(1)
		return s, nil
	case key.Matches(msg, keys.guidePageDown):
		s.prepareLessonGuide()
		s.lessons.guide.ScrollDown(1)
		return s, nil
	default:
		return s.handleTerminalKey(msg)
	}
}

// handleLessonPageKey changes the active lesson page within its bounds and
// returns the guide to the top, reporting whether it consumed the keypress.
func (s *State) handleLessonPageKey(msg tea.KeyPressMsg) bool {
	pages := s.lessons.available[s.lessons.activeIdx].Pages
	switch {
	case key.Matches(msg, keys.previousPage):
		if s.lessons.activePage > 0 {
			s.lessons.activePage--
			s.lessons.guide.GotoTop()
		}
		return true
	case key.Matches(msg, keys.nextPage):
		if s.lessons.activePage < len(pages)-1 {
			s.lessons.activePage++
			s.lessons.guide.GotoTop()
		}
		return true
	default:
		return false
	}
}

// handleTerminalKey forwards input to an active session, records a requested
// exit, or clears a terminal failure while returning to its parent screen.
func (s State) handleTerminalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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
