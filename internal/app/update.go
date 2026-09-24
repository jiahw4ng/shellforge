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
	case LessonsLoadedMsg:
		m.Lessons.Error = msg.Err
		if msg.Err == nil {
			m.Lessons.Available = msg.Lessons
		}
	case CompletionsLoadedMsg:
		if msg.Err != nil {
			slog.Error("Could not load lesson completion", "error", msg.Err)
			break
		}
		if m.Lessons.Completed == nil {
			m.Lessons.Completed = make(map[string]bool)
		}
		for _, lessonID := range msg.LessonIDs {
			m.Lessons.Completed[lessonID] = true
		}
	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		if m.Term.Session != nil {
			return m, m.resizeTerminal()
		}
	case TerminalStartedMsg:
		if msg.Generation != m.Term.generation {
			return m, nil
		}
		m.Term.IsStarting = false
		if msg.Err != nil {
			m.Term.Error = msg.Err
			return m, nil
		}
		if msg.Session == nil {
			m.Term.Error = terminal.NewInvalidStartResultError()
			return m, nil
		}
		m.Term.Session = msg.Session
		m.Term.Output = ""
		m.Term.hasRequestedExit = false
		m.Lessons.Progress = progressState{}
		return m, tea.Batch(m.Term.Session.Init(), waitForTerminalExit(m.Term.Session.Exited(), m.Term.generation), m.resizeTerminal())
	case TerminalExitedMsg:
		if msg.Generation != m.Term.generation {
			return m, nil
		}
		m.Term.IsStarting = false
		m.Lessons.Progress.IsChecking = false
		if !m.Term.hasRequestedExit && m.Term.Session != nil {
			m.Term.Output = m.Term.Session.View()
		}
		m.closeTerminal()
		if m.Term.hasRequestedExit {
			m.Term.Output = ""
			m.Term.Error = nil
			m.returnFromTerminal()
		} else {
			m.Term.Error = errors.New("the terminal closed unexpectedly")
		}
	case AssertionsCheckedMsg:
		if msg.Generation != m.Term.generation {
			return m, nil
		}
		m.Lessons.Progress.IsChecking = false
		m.Lessons.Progress.HasChecked = true
		m.Lessons.Progress.Results = msg.Results
		if msg.LessonID == "" || !hasPassedAllAssertions(msg.Results) {
			break
		}
		if m.Lessons.Completed == nil {
			m.Lessons.Completed = make(map[string]bool)
		}
		if m.Lessons.Completed[msg.LessonID] {
			break
		}
		m.Lessons.Completed[msg.LessonID] = true
		if m.Lessons.Store != nil {
			return m, saveCompletion(m.Lessons.Store, msg.LessonID)
		}
	case CompletionSavedMsg:
		if msg.Err != nil {
			slog.Error("Could not persist lesson completion", "lesson_id", msg.LessonID, "error", msg.Err)
		}
	case CompletionsResetMsg:
		m.Settings.IsResetting = false
		m.Nav.Screen = settingsScreen
		m.Nav.Selection = 0
		if msg.Err != nil {
			m.Settings.Message = "Lesson progress could not be reset."
			m.Settings.Failed = true
			slog.Error("Could not reset lesson completion", "error", msg.Err)
			break
		}
		m.Lessons.Completed = make(map[string]bool)
		m.Lessons.Progress = progressState{}
		m.Settings.Message = "Lesson progress has been reset."
		m.Settings.Failed = false
	case spinner.TickMsg:
		if m.Term.IsStarting {
			updated, command := m.Term.Spinner.Update(msg)
			m.Term.Spinner = updated
			return m, command
		}
	case tea.KeyPressMsg:
		if m.Settings.IsResetting {
			return m, nil
		}
		if m.Nav.Screen == resetConfirmationScreen && msg.Code == tea.KeyEnter && m.Nav.Selection == 1 {
			if m.Lessons.Store == nil {
				m.Nav.Screen = settingsScreen
				m.Nav.Selection = 0
				m.Settings.Message = "Lesson progress could not be reset because storage is unavailable."
				m.Settings.Failed = true
				return m, nil
			}
			m.Settings.IsResetting = true
			return m, resetLessonCompletions(m.Lessons.Store)
		}
		if m.Nav.Screen == lessonScreen && msg.Code == tea.KeyF12 {
			if m.Term.Session == nil || m.Lessons.Progress.IsChecking || m.Lessons.ActiveIndex < 0 || m.Lessons.ActiveIndex >= len(m.Lessons.Available) {
				return m, nil
			}
			m.Lessons.Progress.IsChecking = true
			lesson := m.Lessons.Available[m.Lessons.ActiveIndex]
			return m, checkAssertions(m.Term.Session, lesson.ID, lesson.Assertions, m.Term.generation)
		}
		if m.Nav.Screen == lessonScreen && msg.Code == 'r' && msg.Mod&tea.ModCtrl != 0 && msg.Mod&tea.ModAlt != 0 {
			if m.Term.IsStarting || m.Lessons.ActiveIndex < 0 || m.Lessons.ActiveIndex >= len(m.Lessons.Available) {
				return m, nil
			}
			m.closeTerminal()
			return m, m.startActiveLessonTerminal()
		}
		if m.Nav.Screen == lessonScreen && m.handleLessonPageKey(msg) {
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
			if m.Nav.Screen == lessonScreen {
				return m, m.startActiveLessonTerminal()
			}
			return m, m.startTerminal(nil)
		}
	default:
		if m.usesTerminal() && m.Term.Session != nil {
			return m, m.Term.Session.Update(message)
		}
	}

	return m, nil
}

// startActiveLessonTerminal creates a fresh sandbox for the lesson currently
// displayed beside the terminal.
func (m *State) startActiveLessonTerminal() tea.Cmd {
	lesson := &m.Lessons.Available[m.Lessons.ActiveIndex]
	return m.startTerminal(lesson)
}

// startTerminal clears session-only state and starts a new terminal attempt.
// Every attempt receives a generation so delayed events from an older session
// cannot alter this one.
func (m *State) startTerminal(lesson *lessons.Lesson) tea.Cmd {
	m.Term.generation++
	m.Term.Output = ""
	m.Term.Error = nil
	m.Term.hasRequestedExit = false
	m.Term.IsStarting = true
	m.Lessons.Progress = progressState{}
	m.Term.Spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	width, height := m.terminalDimensions()
	return tea.Batch(startTerminal(width, height, lesson, m.Term.generation), m.Term.Spinner.Tick)
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
	if m.Lessons.ActiveIndex < 0 || m.Lessons.ActiveIndex >= len(m.Lessons.Available) || msg.Mod&tea.ModCtrl == 0 {
		return false
	}

	pages := m.Lessons.Available[m.Lessons.ActiveIndex].Pages
	switch msg.Code {
	case 'p', 'P':
		if m.Lessons.ActivePage > 0 {
			m.Lessons.ActivePage--
		}
		return true
	case 'n', 'N':
		if m.Lessons.ActivePage < len(pages)-1 {
			m.Lessons.ActivePage++
		}
		return true
	default:
		return false
	}
}

func (m State) handleTerminalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.Term.Session != nil {
		if msg.String() == "ctrl+d" {
			m.Term.hasRequestedExit = true
		}
		return m, m.Term.Session.Update(msg)
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if msg.String() == "ctrl+d" && m.Term.Error != nil {
		m.returnFromTerminal()
		m.Term.Error = nil
		m.Term.Output = ""
	}
	return m, nil
}

// returnFromTerminal returns to the screen that launched the terminal.
func (m *State) returnFromTerminal() {
	m.Term.hasRequestedExit = false
	if m.Nav.Screen == lessonScreen {
		m.Nav.Screen = lessonsScreen
		return
	}
	m.Nav.Screen = menuScreen
}

func (m State) usesTerminal() bool {
	return m.Nav.Screen == terminalScreen || m.Nav.Screen == lessonScreen
}

func (m State) resizeTerminal() tea.Cmd {
	width, height := ui.ApplicationContentDimensions(m.Viewport.Width, m.Viewport.Height)
	if m.Nav.Screen == lessonScreen {
		width, height = screens.LessonTerminalDimensions(width, height)
	}
	return m.Term.Session.Resize(width, height)
}

func (m State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(m.Viewport.Width, m.Viewport.Height)
	if m.Nav.Screen == lessonScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return ui.DimensionWithFallback(width, 80), ui.DimensionWithFallback(height, 24)
}

func (m *State) closeTerminal() {
	if m.Term.Session != nil {
		m.Term.Session.Close()
	}
	m.Term.Session = nil
}

// Close releases an active terminal and removes its disposable container.
func (m *State) Close() {
	m.closeTerminal()
}
