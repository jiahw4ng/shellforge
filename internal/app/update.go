package app

import (
	"errors"
	"log/slog"
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"shellforge/internal/screens"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea events and advances the application state.
func (s State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case lessonsLoadedMsg:
		s.lessons.err = msg.err
		if msg.err == nil {
			s.lessons.available = msg.lessons
		}
	case completionsLoadedMsg:
		if msg.err != nil {
			slog.Error("Could not load lesson completion", "error", msg.err)
			break
		}
		if s.lessons.completed == nil {
			s.lessons.completed = make(map[string]bool)
		}
		for _, lessonID := range msg.lessonIDs {
			s.lessons.completed[lessonID] = true
		}
	case tea.WindowSizeMsg:
		s.viewport.width = msg.Width
		s.viewport.height = msg.Height
		s.prepareLessonGuide()
		if s.term.session != nil {
			return s, s.resizeTerminal()
		}
	case termStartedMsg:
		if msg.gen != s.term.gen {
			return s, nil
		}
		s.term.isStarting = false
		if msg.err != nil {
			s.term.err = msg.err
			return s, nil
		}
		if msg.session == nil {
			s.term.err = terminal.NewInvalidStartResultError()
			return s, nil
		}
		s.term.session = msg.session
		s.term.output = ""
		s.term.hasRequestedExit = false
		s.lessons.progress = progressState{}
		return s, tea.Batch(s.term.session.Init(), waitForTerminalExit(s.term.session.Exited(), s.term.gen), s.resizeTerminal())
	case termExitedMsg:
		if msg.gen != s.term.gen {
			return s, nil
		}
		s.term.isStarting = false
		s.lessons.progress.isChecking = false
		if !s.term.hasRequestedExit && s.term.session != nil {
			s.term.output = s.term.session.View()
		}
		s.closeTerminal()
		if s.term.hasRequestedExit {
			s.term.output = ""
			s.term.err = nil
			s.returnFromTerminal()
		} else {
			s.term.err = errors.New("the terminal closed unexpectedly")
		}
	case assertionsCheckedMsg:
		if msg.gen != s.term.gen {
			return s, nil
		}
		s.lessons.progress.isChecking = false
		s.lessons.progress.hasChecked = true
		s.lessons.progress.results = msg.results
		if msg.lessonID == "" || !assertion.HasPassedAllAssertions(msg.results) {
			break
		}
		if s.lessons.completed == nil {
			s.lessons.completed = make(map[string]bool)
		}
		if s.lessons.completed[msg.lessonID] {
			break
		}
		s.lessons.completed[msg.lessonID] = true
		if s.lessons.store != nil {
			return s, saveCompletion(s.lessons.store, msg.lessonID)
		}
	case completionSavedMsg:
		if msg.err != nil {
			slog.Error("Could not persist lesson completion", "lesson_id", msg.lessonID, "error", msg.err)
		}
	case completionsResetMsg:
		s.settings.isResetting = false
		s.nav.screen = settingsScreen
		s.nav.selection = 0
		if msg.err != nil {
			s.settings.message = "Lesson progress could not be reset."
			s.settings.failed = true
			slog.Error("Could not reset lesson completion", "error", msg.err)
			break
		}
		s.lessons.completed = make(map[string]bool)
		s.lessons.progress = progressState{}
		s.settings.message = "Lesson progress has been reset."
		s.settings.failed = false
	case spinner.TickMsg:
		if s.term.isStarting {
			updated, command := s.term.spinner.Update(msg)
			s.term.spinner = updated
			return s, command
		}
	case tea.KeyPressMsg:
		if s.settings.isResetting {
			return s, nil
		}
		if s.nav.screen == resetConfirmationScreen && key.Matches(msg, keys.selectItem) && s.nav.selection == 1 {
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
		if s.nav.screen == lessonScreen && key.Matches(msg, keys.checkProgress) {
			if s.term.session == nil || s.lessons.progress.isChecking || s.lessons.activeIdx < 0 || s.lessons.activeIdx >= len(s.lessons.available) {
				return s, nil
			}
			s.lessons.progress.isChecking = true
			lesson := s.lessons.available[s.lessons.activeIdx]
			return s, checkAssertions(s.term.session, lesson.ID, lesson.Assertions, s.term.gen)
		}
		if s.nav.screen == lessonScreen && key.Matches(msg, keys.resetSandbox) {
			if s.term.isStarting || s.lessons.activeIdx < 0 || s.lessons.activeIdx >= len(s.lessons.available) {
				return s, nil
			}
			s.closeTerminal()
			return s, s.startActiveLessonTerminal()
		}
		if s.nav.screen == lessonScreen && s.handleLessonPageKey(msg) {
			return s, nil
		}
		if s.nav.screen == lessonScreen && (key.Matches(msg, keys.guidePageUp) || key.Matches(msg, keys.guidePageDown)) {
			s.prepareLessonGuide()
			updated, command := s.lessons.guide.Update(msg)
			s.lessons.guide = updated
			return s, command
		}
		// first check: if user is already using the terminal, send keypresses to it
		if s.usesTerminal() {
			return s.handleTerminalKey(msg)
		}
		if s.handleNavigationKey(msg) {
			return s, tea.Quit
		}
		// second check: if the navigation resulted in a screen that uses the terminal, start it
		if s.usesTerminal() {
			if s.nav.screen == lessonScreen {
				return s, s.startActiveLessonTerminal()
			}
			return s, s.startTerminal(nil)
		}
	default:
		if s.usesTerminal() && s.term.session != nil {
			return s, s.term.session.Update(message)
		}
	}

	return s, nil
}

// startActiveLessonTerminal creates a fresh sandbox for the lesson currently
// displayed beside the terminal.
func (s *State) startActiveLessonTerminal() tea.Cmd {
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

// handleLessonPageKey reserves Ctrl+P and Ctrl+N for lesson navigation before
// Bubbleterm can forward those shortcuts to Bash.
// returns true if the keypress was handled, false otherwise.
func (s *State) handleLessonPageKey(msg tea.KeyPressMsg) bool {
	if s.lessons.activeIdx < 0 || s.lessons.activeIdx >= len(s.lessons.available) {
		return false
	}

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

// returnFromTerminal returns to the screen that launched the terminal.
func (s *State) returnFromTerminal() {
	s.term.hasRequestedExit = false
	if s.nav.screen == lessonScreen {
		s.nav.screen = lessonsScreen
		return
	}
	s.nav.screen = menuScreen
}

func (s State) usesTerminal() bool {
	return s.nav.screen == terminalScreen || s.nav.screen == lessonScreen
}

func (s State) resizeTerminal() tea.Cmd {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	if s.nav.screen == lessonScreen {
		width, height = screens.LessonTerminalDimensions(width, height)
	}
	return s.term.session.Resize(width, height)
}

func (s State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	if s.nav.screen == lessonScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return ui.DimensionWithFallback(width, 80), ui.DimensionWithFallback(height, 24)
}

func (s *State) closeTerminal() {
	if s.term.session != nil {
		s.term.session.Close()
	}
	s.term.session = nil
}

// Close releases an active terminal and removes its disposable container.
func (s *State) Close() {
	s.closeTerminal()
}
