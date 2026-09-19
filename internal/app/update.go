package app

import (
	"errors"
	"shellforge/internal/lessons"
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea events and advances the application state.
func (m State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case LessonsLoadedMsg:
		m.LessonErr = msg.Err
		if msg.Err == nil {
			m.Lessons = msg.Lessons
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.Terminal != nil {
			return m, m.resizeTerminal()
		}
	case TerminalStartedMsg:
		m.TerminalErr = msg.Err
		if msg.Session != nil {
			if msg.Session == nil && msg.Err == nil {
				m.TerminalErr = errors.New("terminal startup returned no session")
				return m, nil
			}
			m.Terminal = msg.Session
			m.TerminalOutput = ""
			m.TerminalExitRequested = false
			return m, tea.Batch(m.Terminal.Init(), waitForTerminalExit(m.Terminal.Exited()), m.resizeTerminal())
		}
	case TerminalExitedMsg:
		if !m.TerminalExitRequested && m.Terminal != nil {
			m.TerminalOutput = m.Terminal.View()
		}
		m.closeTerminal()
		if m.TerminalExitRequested {
			m.TerminalOutput = ""
			m.TerminalErr = nil
			m.returnFromTerminal()
		} else {
			m.TerminalErr = errors.New("the terminal closed unexpectedly")
		}
	case tea.KeyPressMsg:
		if m.CurrentScreen == lessonScreen && m.handleLessonPageKey(msg) {
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
			if m.CurrentScreen == lessonScreen {
				selectedLesson := m.ActiveLesson
				lessonToBuild = &m.Lessons[selectedLesson]
			}
			// A previous failed session may have preserved its final frame. Clear it
			// before the asynchronous startup command runs so it cannot appear in
			// the new terminal pane.
			m.TerminalOutput = ""
			m.TerminalErr = nil
			m.TerminalExitRequested = false
			width, height := m.terminalDimensions()
			return m, startTerminal(width, height, lessonToBuild)
		}
	default:
		if m.usesTerminal() && m.Terminal != nil {
			return m, m.Terminal.Update(message)
		}
	}

	return m, nil
}

// handleLessonPageKey reserves Ctrl+[ and Ctrl+] for lesson navigation before
// Bubbleterm can forward those shortcuts to Bash.
// returns true if the keypress was handled, false otherwise.
func (m *State) handleLessonPageKey(msg tea.KeyPressMsg) bool {
	if m.ActiveLesson < 0 || m.ActiveLesson >= len(m.Lessons) || msg.Mod&tea.ModCtrl == 0 {
		return false
	}

	pages := m.Lessons[m.ActiveLesson].Pages
	switch {
	case msg.Code == '[' || msg.ShiftedCode == ']':
		if m.ActivePage > 0 {
			m.ActivePage--
		}
		return true
	case msg.Code == ']' || msg.ShiftedCode == '[':
		if m.ActivePage < len(pages)-1 {
			m.ActivePage++
		}
		return true
	default:
		return false
	}
}

func (m State) handleTerminalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.Terminal != nil {
		if msg.String() == "ctrl+d" {
			m.TerminalExitRequested = true
		}
		return m, m.Terminal.Update(msg)
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if msg.String() == "ctrl+d" && m.TerminalErr != nil {
		m.returnFromTerminal()
		m.TerminalErr = nil
		m.TerminalOutput = ""
	}
	return m, nil
}

// returnFromTerminal returns to the screen that launched the terminal.
func (m *State) returnFromTerminal() {
	m.TerminalExitRequested = false
	if m.CurrentScreen == lessonScreen {
		m.CurrentScreen = lessonsScreen
		return
	}
	m.CurrentScreen = menuScreen
}

func (m State) usesTerminal() bool {
	return m.CurrentScreen == terminalScreen || m.CurrentScreen == lessonScreen
}

func (m State) resizeTerminal() tea.Cmd {
	width, height := ui.ApplicationContentDimensions(m.Width, m.Height)
	if m.CurrentScreen == lessonScreen {
		width, height = screens.LessonTerminalDimensions(width, height)
	}
	return m.Terminal.Resize(width, height)
}

func (m State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(m.Width, m.Height)
	if m.CurrentScreen == lessonScreen {
		return screens.LessonTerminalDimensions(width, height)
	}
	return dimension(width, 80), dimension(height, 24)
}

func dimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func (m *State) closeTerminal() {
	if m.Terminal != nil {
		m.Terminal.Close()
	}
	m.Terminal = nil
}

// Close releases an active terminal and removes its disposable container.
func (m *State) Close() {
	m.closeTerminal()
}
