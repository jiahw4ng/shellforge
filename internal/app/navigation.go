package app

import tea "charm.land/bubbletea/v2"

// handleNavigationKey updates non-terminal navigation and reports whether the app should quit.
func (m *State) handleNavigationKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.nav.screen == menuScreen && m.nav.selection > 0 {
			m.nav.selection--
		}
		if m.nav.screen == lessonsScreen && m.nav.selection > 0 {
			m.nav.selection--
		}
		if (m.nav.screen == settingsScreen || m.nav.screen == resetConfirmationScreen) && m.nav.selection > 0 {
			m.nav.selection--
		}
	case "down":
		if m.nav.screen == menuScreen && m.nav.selection < len(menuItems)-1 {
			m.nav.selection++
		}
		if m.nav.screen == lessonsScreen && m.nav.selection < len(m.lessons.available) {
			m.nav.selection++
		}
		if m.nav.screen == settingsScreen && m.nav.selection < len(settingsItems)-1 {
			m.nav.selection++
		}
		if m.nav.screen == resetConfirmationScreen && m.nav.selection < len(resetConfirmationItems)-1 {
			m.nav.selection++
		}
	case "enter":
		if m.nav.screen == menuScreen && m.nav.selection == len(menuItems)-1 {
			return true
		}
		m.handleEnter()
	}

	return false
}

// handleEnter updates the application state when the user presses the Enter key.
func (m *State) handleEnter() {
	switch m.nav.screen {
	case menuScreen:
		switch m.nav.selection {
		case 0:
			m.nav.screen = terminalScreen
		case 1:
			m.nav.screen = lessonsScreen
			m.nav.selection = 0
		case 2:
			m.nav.screen = settingsScreen
			m.nav.selection = 0
			m.settings = settingsState{}
		}
	case lessonsScreen:
		if len(m.lessons.available) == 0 {
			if m.lessons.err != nil {
				m.nav.screen = menuScreen
				m.nav.selection = 0
			}
			return
		}
		if m.nav.selection == len(m.lessons.available) {
			m.nav.screen = menuScreen
			m.nav.selection = 0
			return
		}
		m.lessons.activeIdx = m.nav.selection
		m.lessons.activePage = 0
		m.nav.screen = lessonScreen
	case settingsScreen:
		if m.nav.selection == len(settingsItems)-1 {
			m.nav.screen = menuScreen
			m.nav.selection = 0
			return
		}
		m.nav.screen = resetConfirmationScreen
		m.nav.selection = 0
	case resetConfirmationScreen:
		if m.nav.selection == 0 {
			m.nav.screen = settingsScreen
			m.nav.selection = 0
		}
	default:
		m.nav.screen = menuScreen
	}
}
