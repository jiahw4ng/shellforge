package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation updates non-terminal navigation and reports whether the app should quit.
func (s *State) handleNavigation(msg tea.KeyMsg) bool {
	switch {
	case key.Matches(msg, keys.quit):
		return true
	case key.Matches(msg, keys.up):
		if s.nav.screen == menuScreen && !outOfBounds(s.nav.selection-1, len(menuItems)) {
			s.nav.selection--
		}
		if s.nav.screen == lessonsScreen && !outOfBounds(s.nav.selection-1, len(s.lessons.available)+1) {
			s.nav.selection--
		}
		if s.nav.screen == settingsScreen && !outOfBounds(s.nav.selection-1, len(settingsItems)) {
			s.nav.selection--
		}
		if s.nav.screen == resetConfirmationScreen && !outOfBounds(s.nav.selection-1, len(resetConfirmationItems)) {
			s.nav.selection--
		}
	case key.Matches(msg, keys.down):
		if s.nav.screen == menuScreen && !outOfBounds(s.nav.selection+1, len(menuItems)) {
			s.nav.selection++
		}
		if s.nav.screen == lessonsScreen && !outOfBounds(s.nav.selection+1, len(s.lessons.available)+1) {
			s.nav.selection++
		}
		if s.nav.screen == settingsScreen && !outOfBounds(s.nav.selection+1, len(settingsItems)) {
			s.nav.selection++
		}
		if s.nav.screen == resetConfirmationScreen && !outOfBounds(s.nav.selection+1, len(resetConfirmationItems)) {
			s.nav.selection++
		}
	case key.Matches(msg, keys.selectItem):
		// if user pressed Enter on the last menu item (Exit), quit the app
		if s.nav.screen == menuScreen && s.nav.selection == len(menuItems)-1 {
			return true
		}
		// else, handle the Enter key for the current screen
		s.handleEnter()
	}

	return false
}

func outOfBounds(idx, length int) bool {
	return idx < 0 || idx >= length
}

// handleEnter updates the application state when the user presses the Enter key.
func (s *State) handleEnter() {
	switch s.nav.screen {
	case menuScreen:
		switch s.nav.selection {
		case 0:
			s.nav.screen = sandboxTerminalScreen
		case 1:
			s.nav.screen = lessonsScreen
			s.nav.selection = 0
		case 2:
			s.nav.screen = settingsScreen
			s.nav.selection = 0
			s.settings = settingsState{}
		}
	case lessonsScreen:
		if len(s.lessons.available) == 0 {
			if s.lessons.err != nil {
				s.nav.screen = menuScreen
				s.nav.selection = 0
			}
			return
		}
		if s.nav.selection == len(s.lessons.available) {
			s.nav.screen = menuScreen
			s.nav.selection = 0
			return
		}
		s.lessons.activeIdx = s.nav.selection
		s.lessons.activePage = 0
		s.lessons.revealedHints = 0
		s.lessons.guide.GotoTop()
		s.nav.screen = lessonTerminalScreen
	case settingsScreen:
		if s.nav.selection == len(settingsItems)-1 {
			s.nav.screen = menuScreen
			s.nav.selection = 0
			return
		}
		s.nav.screen = resetConfirmationScreen
		s.nav.selection = 0
	case resetConfirmationScreen:
		if s.nav.selection == 0 {
			s.nav.screen = settingsScreen
			s.nav.selection = 0
		}
	default:
		s.nav.screen = menuScreen
	}
}
