package app

import (
	"errors"
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
	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		if m.Term.Session != nil {
			return m, m.resizeTerminal()
		}
	case TerminalStartedMsg:
		m.Term.isStarting = false
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
		return m, tea.Batch(m.Term.Session.Init(), waitForTerminalExit(m.Term.Session.Exited()), m.resizeTerminal())
	case TerminalExitedMsg:
		m.Term.isStarting = false
		m.Lessons.Progress.isChecking = false
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
		m.Lessons.Progress.isChecking = false
		m.Lessons.Progress.hasChecked = true
		m.Lessons.Progress.Results = msg.Results
	case spinner.TickMsg:
		if m.Term.isStarting {
			updated, command := m.Term.Spinner.Update(msg)
			m.Term.Spinner = updated
			return m, command
		}
	case tea.KeyPressMsg:
		if m.Nav.Screen == lessonScreen && msg.Code == tea.KeyF12 {
			if m.Term.Session == nil || m.Lessons.Progress.isChecking || m.Lessons.ActiveIndex < 0 || m.Lessons.ActiveIndex >= len(m.Lessons.Available) {
				return m, nil
			}
			m.Lessons.Progress.isChecking = true
			return m, checkAssertions(m.Term.Session, m.Lessons.Available[m.Lessons.ActiveIndex].Assertions)
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
			// if the user is currently on the lesson screen, pass the selected lesson to the terminal
			// so it can build the sandbox filestructure correctly
			var lessonToBuild *lessons.Lesson
			if m.Nav.Screen == lessonScreen {
				selectedLesson := m.Lessons.ActiveIndex
				lessonToBuild = &m.Lessons.Available[selectedLesson]
			}
			// A previous failed session may have preserved its final frame. Clear it
			// before the asynchronous startup command runs so it cannot appear in
			// the new terminal pane.
			m.Term.Output = ""
			m.Term.Error = nil
			m.Term.hasRequestedExit = false
			m.Term.isStarting = true
			m.Lessons.Progress = progressState{}
			m.Term.Spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
			width, height := m.terminalDimensions()
			return m, tea.Batch(startTerminal(width, height, lessonToBuild), m.Term.Spinner.Tick)
		}
	default:
		if m.usesTerminal() && m.Term.Session != nil {
			return m, m.Term.Session.Update(message)
		}
	}

	return m, nil
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
