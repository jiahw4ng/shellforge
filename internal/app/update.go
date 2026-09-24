package app

import (
	"errors"
	"log/slog"
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"shellforge/internal/screens"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea events and advances the application state.
func (m State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case lessonsLoadedMsg:
		m.lessons.err = msg.err
		if msg.err == nil {
			m.lessons.available = msg.lessons
		}
	case completionsLoadedMsg:
		if msg.err != nil {
			slog.Error("Could not load lesson completion", "error", msg.err)
			break
		}
		if m.lessons.completed == nil {
			m.lessons.completed = make(map[string]bool)
		}
		for _, lessonID := range msg.lessonIDs {
			m.lessons.completed[lessonID] = true
		}
	case tea.WindowSizeMsg:
		m.viewport.width = msg.Width
		m.viewport.height = msg.Height
		if m.term.session != nil {
			return m, m.resizeTerminal()
		}
	case termStartedMsg:
		if msg.gen != m.term.generation {
			return m, nil
		}
		m.term.isStarting = false
		if msg.err != nil {
			m.term.err = msg.err
			return m, nil
		}
		if msg.session == nil {
			m.term.err = terminal.NewInvalidStartResultError()
			return m, nil
		}
		m.term.session = msg.session
		m.term.output = ""
		m.term.hasRequestedExit = false
		m.lessons.progress = progressState{}
		return m, tea.Batch(m.term.session.Init(), waitForTerminalExit(m.term.session.Exited(), m.term.generation), m.resizeTerminal())
	case termExitedMsg:
		if msg.gen != m.term.generation {
			return m, nil
		}
		m.term.isStarting = false
		m.lessons.progress.isChecking = false
		if !m.term.hasRequestedExit && m.term.session != nil {
			m.term.output = m.term.session.View()
		}
		m.closeTerminal()
		if m.term.hasRequestedExit {
			m.term.output = ""
			m.term.err = nil
			m.returnFromTerminal()
		} else {
			m.term.err = errors.New("the terminal closed unexpectedly")
		}
	case assertionsCheckedMsg:
		if msg.gen != m.term.generation {
			return m, nil
		}
		m.lessons.progress.isChecking = false
		m.lessons.progress.hasChecked = true
		m.lessons.progress.results = msg.results
		if msg.lessonID == "" || !hasPassedAllAssertions(msg.results) {
			break
		}
		if m.lessons.completed == nil {
			m.lessons.completed = make(map[string]bool)
		}
		if m.lessons.completed[msg.lessonID] {
			break
		}
		m.lessons.completed[msg.lessonID] = true
		if m.lessons.store != nil {
			return m, saveCompletion(m.lessons.store, msg.lessonID)
		}
	case completionSavedMsg:
		if msg.err != nil {
			slog.Error("Could not persist lesson completion", "lesson_id", msg.lessonID, "error", msg.err)
		}
	case completionsResetMsg:
		m.settings.isResetting = false
		m.nav.screen = settingsScreen
		m.nav.selection = 0
		if msg.err != nil {
			m.settings.message = "Lesson progress could not be reset."
			m.settings.failed = true
			slog.Error("Could not reset lesson completion", "error", msg.err)
			break
		}
		m.lessons.completed = make(map[string]bool)
		m.lessons.progress = progressState{}
		m.settings.message = "Lesson progress has been reset."
		m.settings.failed = false
	case spinner.TickMsg:
		if m.term.isStarting {
			updated, command := m.term.spinner.Update(msg)
			m.term.spinner = updated
			return m, command
		}
	case tea.KeyPressMsg:
		if m.settings.isResetting {
			return m, nil
		}
		if m.nav.screen == resetConfirmationScreen && msg.Code == tea.KeyEnter && m.nav.selection == 1 {
			if m.lessons.store == nil {
				m.nav.screen = settingsScreen
				m.nav.selection = 0
				m.settings.message = "Lesson progress could not be reset because storage is unavailable."
				m.settings.failed = true
				return m, nil
			}
			m.settings.isResetting = true
			return m, resetLessonCompletions(m.lessons.store)
		}
		if m.nav.screen == lessonScreen && msg.Code == tea.KeyF12 {
			if m.term.session == nil || m.lessons.progress.isChecking || m.lessons.activeIdx < 0 || m.lessons.activeIdx >= len(m.lessons.available) {
				return m, nil
			}
			m.lessons.progress.isChecking = true
			lesson := m.lessons.available[m.lessons.activeIdx]
			return m, checkAssertions(m.term.session, lesson.ID, lesson.Assertions, m.term.generation)
		}
		if m.nav.screen == lessonScreen && msg.Code == 'r' && msg.Mod&tea.ModCtrl != 0 && msg.Mod&tea.ModAlt != 0 {
			if m.term.isStarting || m.lessons.activeIdx < 0 || m.lessons.activeIdx >= len(m.lessons.available) {
				return m, nil
			}
			m.closeTerminal()
			return m, m.startActiveLessonTerminal()
		}
		if m.nav.screen == lessonScreen && m.handleLessonPageKey(msg) {
			return m, nil
		}
		// first check: if user is already using the terminal, send keypresses to it
		if m.usesTerminal() {
			return m.handleTerminalKey(msg)
		}
		if m.handleNavigationKey(msg) {
			return m, tea.Quit
		}
		// second check: if the navigation resulted in a screen that uses the terminal, start it
		if m.usesTerminal() {
			if m.nav.screen == lessonScreen {
				return m, m.startActiveLessonTerminal()
			}
			return m, m.startTerminal(nil)
		}
	default:
		if m.usesTerminal() && m.term.session != nil {
			return m, m.term.session.Update(message)
		}
	}

	return m, nil
}

// startActiveLessonTerminal creates a fresh sandbox for the lesson currently
// displayed beside the terminal.
func (m *State) startActiveLessonTerminal() tea.Cmd {
	lesson := &m.lessons.available[m.lessons.activeIdx]
	return m.startTerminal(lesson)
}

// startTerminal clears session-only state and starts a new terminal attempt.
// Every attempt receives a generation so delayed events from an older session
// cannot alter this one.
func (m *State) startTerminal(lesson *lessons.Lesson) tea.Cmd {
	m.term.generation++
	m.term.output = ""
	m.term.err = nil
	m.term.hasRequestedExit = false
	m.term.isStarting = true
	m.lessons.progress = progressState{}
	m.term.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	width, height := m.terminalDimensions()
	return tea.Batch(startTerminal(width, height, lesson, m.term.generation), m.term.spinner.Tick)
}

func hasPassedAllAssertions(results []assertion.Result) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

// handleLessonPageKey reserves Ctrl+P and Ctrl+N for lesson navigation before
// Bubbleterm can forward those shortcuts to Bash.
// returns true if the keypress was handled, false otherwise.
func (m *State) handleLessonPageKey(msg tea.KeyPressMsg) bool {
	if m.lessons.activeIdx < 0 || m.lessons.activeIdx >= len(m.lessons.available) || msg.Mod&tea.ModCtrl == 0 {
		return false
	}

	pages := m.lessons.available[m.lessons.activeIdx].Pages
	switch msg.Code {
	case 'p', 'P':
		if m.lessons.activePage > 0 {
			m.lessons.activePage--
		}
		return true
	case 'n', 'N':
		if m.lessons.activePage < len(pages)-1 {
			m.lessons.activePage++
		}
		return true
	default:
		return false
	}
}

func (m State) handleTerminalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.term.session != nil {
		if msg.String() == "ctrl+d" {
			m.term.hasRequestedExit = true
		}
		return m, m.term.session.Update(msg)
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if msg.String() == "ctrl+d" && m.term.err != nil {
		m.returnFromTerminal()
		m.term.err = nil
		m.term.output = ""
	}
	return m, nil
}

// returnFromTerminal returns to the screen that launched the terminal.
func (m *State) returnFromTerminal() {
	m.term.hasRequestedExit = false
	if m.nav.screen == lessonScreen {
		m.nav.screen = lessonsScreen
		return
	}
	m.nav.screen = menuScreen
}

func (m State) usesTerminal() bool {
	return m.nav.screen == terminalScreen || m.nav.screen == lessonScreen
}

func (m State) resizeTerminal() tea.Cmd {
	width, height := ui.ApplicationContentDimensions(m.viewport.width, m.viewport.height)
	if m.nav.screen == lessonScreen {
		width, height = screens.LessonTerminalDimensions(width, height)
	}
	return m.term.session.Resize(width, height)
}

func (m State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(m.viewport.width, m.viewport.height)
	if m.nav.screen == lessonScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return ui.DimensionWithFallback(width, 80), ui.DimensionWithFallback(height, 24)
}

func (m *State) closeTerminal() {
	if m.term.session != nil {
		m.term.session.Close()
	}
	m.term.session = nil
}

// Close releases an active terminal and removes its disposable container.
func (m *State) Close() {
	m.closeTerminal()
}
