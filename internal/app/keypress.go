package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress leaves state unchanged while a reset is running and otherwise
// delegates the keypress to the handler for the active screen.
func (s State) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.settings.isResetting {
		return s, nil
	}
	if s.nav.screen != config.SandboxTerminalScreen && key.Matches(msg, config.Keys.ToggleHelp) {
		s.help.ShowAll = !s.help.ShowAll
		if s.nav.screen == config.LessonTerminalScreen {
			s.prepareLessonGuide()
		}
		return s, nil
	}

	switch s.nav.screen {
	case config.LessonTerminalScreen:
		return s.handleLessonKeyPress(msg)
	case config.SandboxTerminalScreen:
		return s.handleTerminalKeyPress(msg)
	case config.ResetConfirmationScreen:
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
	case config.LessonTerminalScreen:
		return s, s.startTerminalWithActiveLesson()
	case config.SandboxTerminalScreen:
		return s, s.startTerminal(nil)
	default:
		return s, nil
	}
}

// handleResetConfirmationKeyPress delegates ordinary navigation, reports unavailable
// storage, or marks reset as active and schedules completion deletion.
func (s State) handleResetConfirmationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// if the user did not press Enter on the correct option, handle navigation normally
	if !key.Matches(msg, config.Keys.SelectItem) || s.nav.selection != 1 {
		return s.handleNavigationKeyPress(msg)
	}

	// guard againt unavailable storage
	if s.lessons.store == nil {
		s.nav.screen = config.SettingsScreen
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
	case key.Matches(msg, config.Keys.ToggleHints):
		s.toggleHints()
		return s, nil
	case key.Matches(msg, config.Keys.ToggleAssertions):
		s.toggleAssertions()
		return s, nil
	case key.Matches(msg, config.Keys.RevealHint):
		s.revealNextHint()
		return s, nil
	case key.Matches(msg, config.Keys.CheckProgress):
		return s, s.startProgressCheck()
	case key.Matches(msg, config.Keys.ResetSandbox):
		return s, s.restartLessonSandbox()
	case key.Matches(msg, config.Keys.PreviousPage):
		s.showPreviousLessonPage()
		return s, nil
	case key.Matches(msg, config.Keys.NextPage):
		s.showNextLessonPage()
		return s, nil
	case key.Matches(msg, config.Keys.GuidePageUp):
		s.scrollLessonGuideUp()
		return s, nil
	case key.Matches(msg, config.Keys.GuidePageDown):
		s.scrollLessonGuideDown()
		return s, nil
	default:
		return s.handleTerminalKeyPress(msg)
	}
}

func (s *State) startProgressCheck() tea.Cmd {
	if s.term.session == nil || s.lessons.progress.isChecking {
		return nil
	}
	s.lessons.progress.isChecking = true
	s.lessons.progress.assertionsExpanded = true
	lesson := s.lessons.available[s.lessons.activeIdx]
	return checkAssertions(s.term.session, lesson.ID, lesson.Assertions, s.term.gen)
}

func (s *State) restartLessonSandbox() tea.Cmd {
	if s.term.isStarting {
		return nil
	}
	s.CloseTerminal()
	return s.startTerminalWithActiveLesson()
}

func (s *State) revealNextHint() {
	hints := s.lessons.available[s.lessons.activeIdx].Hints
	if s.lessons.revealedHints >= len(hints) {
		return
	}
	s.lessons.revealedHints++
	s.lessons.hintsExpanded = true
	s.prepareLessonGuide()
}

func (s *State) toggleHints() {
	s.lessons.hintsExpanded = !s.lessons.hintsExpanded
	s.prepareLessonGuide()
}

func (s *State) toggleAssertions() {
	s.lessons.progress.assertionsExpanded = !s.lessons.progress.assertionsExpanded
	s.prepareLessonGuide()
}

func (s *State) scrollLessonGuideUp() {
	s.prepareLessonGuide()
	s.lessons.guide.ScrollUp(1)
}

func (s *State) scrollLessonGuideDown() {
	s.prepareLessonGuide()
	s.lessons.guide.ScrollDown(1)
}

func (s *State) showPreviousLessonPage() {
	if s.lessons.activePage > 0 {
		s.lessons.activePage--
		s.help.ShowAll = false
		s.lessons.guide.GotoTop()

	}
}

func (s *State) showNextLessonPage() {
	pages := s.lessons.available[s.lessons.activeIdx].Pages
	if s.lessons.activePage < len(pages)-1 {
		s.lessons.activePage++
		s.help.ShowAll = false
		s.lessons.guide.GotoTop()
	}
}

// handleTerminalKeyPress forwards input to an active session, records a requested
// exit, or clears a terminal failure while returning to its parent screen.
// note: this is used by both the lesson and standalone sandbox terminal.
func (s State) handleTerminalKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.term.session != nil {
		if key.Matches(msg, config.Keys.ReturnBack) {
			s.term.hasRequestedExit = true
		}
		return s, s.term.session.Update(msg)
	}
	if key.Matches(msg, config.Keys.Quit) {
		return s, tea.Quit
	}
	if key.Matches(msg, config.Keys.ReturnBack) && s.term.err != nil {
		s.returnFromTerminal()
		s.term.err = nil
		s.term.output = ""
	}
	return s, nil
}
