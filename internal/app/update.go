package app

import (
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
			m.Terminal = msg.Session
			return m, tea.Batch(m.Terminal.Init(), waitForTerminalExit(m.Terminal.Exited()), m.resizeTerminal())
		}
	case TerminalExitedMsg:
		m.closeTerminal()
		if m.CurrentScreen == lessonScreen {
			m.CurrentScreen = lessonsScreen
		} else {
			m.CurrentScreen = menuScreen
		}
	case tea.KeyPressMsg:
		if m.usesTerminal() {
			return m.handleTerminalKey(msg)
		}
		if m.handleNavigationKey(msg) {
			return m, tea.Quit
		}
		if m.usesTerminal() {
			width, height := m.terminalDimensions()
			return m, startTerminal(width, height)
		}
	default:
		if m.usesTerminal() && m.Terminal != nil {
			return m, m.Terminal.Update(message)
		}
	}

	return m, nil
}

func (m State) handleTerminalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.Terminal != nil {
		return m, m.Terminal.Update(msg)
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if msg.String() == "enter" && m.TerminalErr != nil {
		if m.CurrentScreen == lessonScreen {
			m.CurrentScreen = lessonsScreen
		} else {
			m.CurrentScreen = menuScreen
		}
		m.TerminalErr = nil
	}
	return m, nil
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
