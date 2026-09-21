package app

import tea "charm.land/bubbletea/v2"

// handleNavigationKey updates non-terminal navigation and reports whether the app should quit.
func (m *State) handleNavigationKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.Nav.Screen == menuScreen && m.Nav.Selection > 0 {
			m.Nav.Selection--
		}
		if m.Nav.Screen == lessonsScreen && m.Nav.Selection > 0 {
			m.Nav.Selection--
		}
		if (m.Nav.Screen == settingsScreen || m.Nav.Screen == resetConfirmationScreen) && m.Nav.Selection > 0 {
			m.Nav.Selection--
		}
	case "down":
		if m.Nav.Screen == menuScreen && m.Nav.Selection < len(menuItems)-1 {
			m.Nav.Selection++
		}
		if m.Nav.Screen == lessonsScreen && m.Nav.Selection < len(m.Lessons.Available) {
			m.Nav.Selection++
		}
		if m.Nav.Screen == settingsScreen && m.Nav.Selection < len(settingsItems)-1 {
			m.Nav.Selection++
		}
		if m.Nav.Screen == resetConfirmationScreen && m.Nav.Selection < len(resetConfirmationItems)-1 {
			m.Nav.Selection++
		}
	case "enter":
		if m.Nav.Screen == menuScreen && m.Nav.Selection == len(menuItems)-1 {
			return true
		}
		m.handleEnter()
	}

	return false
}

// handleEnter updates the application state when the user presses the Enter key.
func (m *State) handleEnter() {
	switch m.Nav.Screen {
	case menuScreen:
		switch m.Nav.Selection {
		case 0:
			m.Nav.Screen = terminalScreen
		case 1:
			m.Nav.Screen = lessonsScreen
			m.Nav.Selection = 0
		case 2:
			m.Nav.Screen = settingsScreen
			m.Nav.Selection = 0
			m.Settings = settingsState{}
		}
	case lessonsScreen:
		if len(m.Lessons.Available) == 0 {
			if m.Lessons.Error != nil {
				m.Nav.Screen = menuScreen
				m.Nav.Selection = 0
			}
			return
		}
		if m.Nav.Selection == len(m.Lessons.Available) {
			m.Nav.Screen = menuScreen
			m.Nav.Selection = 0
			return
		}
		m.Lessons.ActiveIndex = m.Nav.Selection
		m.Lessons.ActivePage = 0
		m.Nav.Screen = lessonScreen
	case settingsScreen:
		if m.Nav.Selection == len(settingsItems)-1 {
			m.Nav.Screen = menuScreen
			m.Nav.Selection = 0
			return
		}
		m.Nav.Screen = resetConfirmationScreen
		m.Nav.Selection = 0
	case resetConfirmationScreen:
		if m.Nav.Selection == 0 {
			m.Nav.Screen = settingsScreen
			m.Nav.Selection = 0
		}
	default:
		m.Nav.Screen = menuScreen
	}
}
