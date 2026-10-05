package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation updates non-terminal navigation and reports whether the app should quit.
func (s *State) handleNavigation(msg tea.KeyMsg) bool {
	switch {
	case key.Matches(msg, config.Keys.Quit):
		return true
	case key.Matches(msg, config.Keys.Up):
		if s.nav.screen == config.MenuScreen && !outOfBounds(s.nav.selection-1, len(config.MenuItems)) {
			s.nav.selection--
		}
		if s.nav.screen == config.LessonsScreen && !outOfBounds(s.nav.selection-1, len(s.lessons.available)+1) {
			s.nav.selection--
		}
		if s.nav.screen == config.SettingsScreen && !outOfBounds(s.nav.selection-1, len(config.SettingsItems)) {
			s.nav.selection--
		}
		if s.nav.screen == config.ResetConfirmationScreen && !outOfBounds(s.nav.selection-1, len(config.ResetConfirmationItems)) {
			s.nav.selection--
		}
	case key.Matches(msg, config.Keys.Down):
		if s.nav.screen == config.MenuScreen && !outOfBounds(s.nav.selection+1, len(config.MenuItems)) {
			s.nav.selection++
		}
		if s.nav.screen == config.LessonsScreen && !outOfBounds(s.nav.selection+1, len(s.lessons.available)+1) {
			s.nav.selection++
		}
		if s.nav.screen == config.SettingsScreen && !outOfBounds(s.nav.selection+1, len(config.SettingsItems)) {
			s.nav.selection++
		}
		if s.nav.screen == config.ResetConfirmationScreen && !outOfBounds(s.nav.selection+1, len(config.ResetConfirmationItems)) {
			s.nav.selection++
		}
	case key.Matches(msg, config.Keys.SelectItem):
		// if user pressed Enter on the last menu item (Exit), quit the app
		if s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1 {
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
	case config.MenuScreen:
		switch s.nav.selection {
		case 0:
			s.nav.screen = config.SandboxTerminalScreen
		case 1:
			s.nav.screen = config.LessonsScreen
			s.nav.selection = 0
		case 2:
			s.nav.screen = config.SettingsScreen
			s.nav.selection = 0
			s.settings = settingsState{}
		}
	case config.LessonsScreen:
		if len(s.lessons.available) == 0 {
			if s.lessons.err != nil {
				s.nav.screen = config.MenuScreen
				s.nav.selection = 0
			}
			return
		}
		if s.nav.selection == len(s.lessons.available) {
			s.nav.screen = config.MenuScreen
			s.nav.selection = 0
			return
		}
		s.lessons.activeIdx = s.nav.selection
		s.lessons.activePage = 0
		s.lessons.revealedHints = 0
		s.lessons.guide.GotoTop()
		s.nav.screen = config.LessonTerminalScreen
	case config.SettingsScreen:
		if s.nav.selection == len(config.SettingsItems)-1 {
			s.nav.screen = config.MenuScreen
			s.nav.selection = 0
			return
		}
		s.nav.screen = config.ResetConfirmationScreen
		s.nav.selection = 0
	case config.ResetConfirmationScreen:
		if s.nav.selection == 0 {
			s.nav.screen = config.SettingsScreen
			s.nav.selection = 0
		}
	default:
		s.nav.screen = config.MenuScreen
	}
}
