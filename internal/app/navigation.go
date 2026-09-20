package app

import tea "charm.land/bubbletea/v2"

// handleNavigationKey updates non-terminal navigation and reports whether the app should quit.
func (m *State) handleNavigationKey(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.CurrentScreen == menuScreen && m.SelectedOption > 0 {
			m.SelectedOption--
		}
		if m.CurrentScreen == lessonsScreen && m.SelectedLesson > 0 {
			m.SelectedLesson--
		}
	case "down":
		if m.CurrentScreen == menuScreen && m.SelectedOption < len(menuItems)-1 {
			m.SelectedOption++
		}
		if m.CurrentScreen == lessonsScreen && m.SelectedLesson < len(m.Lessons) {
			m.SelectedLesson++
		}
	case "enter":
		if m.CurrentScreen == menuScreen && m.SelectedOption == len(menuItems)-1 {
			return true
		}
		m.handleEnter()
	}

	return false
}

// handleEnter updates the application state when the user presses the Enter key.
func (m *State) handleEnter() {
	switch m.CurrentScreen {
	case menuScreen:
		switch m.SelectedOption {
		case 0:
			m.CurrentScreen = terminalScreen
		case 1:
			m.CurrentScreen = lessonsScreen
		default:
			m.CurrentScreen = featureScreen
			m.FeatureReturnTo = menuScreen
		}
	case lessonsScreen:
		if len(m.Lessons) == 0 {
			if m.LessonErr != nil {
				m.CurrentScreen = menuScreen
			}
			return
		}
		if m.SelectedLesson == len(m.Lessons) {
			m.CurrentScreen = menuScreen
			return
		}
		m.ActiveLesson = m.SelectedLesson
		m.ActivePage = 0
		m.CurrentScreen = lessonScreen
	case featureScreen:
		m.CurrentScreen = m.FeatureReturnTo
	default:
		m.CurrentScreen = menuScreen
	}
}
