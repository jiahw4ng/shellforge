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
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		if s.exitSelected() {
			return true
		}
		s.handleEnter()
	}

	return false
}

func (s State) exitSelected() bool {
	return s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1
}

func (s State) selectableItemCount() int {
	switch s.nav.screen {
	case config.MenuScreen:
		return len(config.MenuItems)
	case config.LessonsScreen:
		if len(s.lessons.available) == 0 && s.lessons.err == nil {
			return 0
		}
		return len(s.lessons.available) + 1
	case config.SettingsScreen:
		return len(config.SettingsItems)
	case config.ResetConfirmationScreen:
		return len(config.ResetConfirmationItems)
	default:
		return 0
	}
}

func (s *State) moveSelection(delta int) {
	next := s.nav.selection + delta
	if next >= 0 && next < s.selectableItemCount() {
		s.nav.selection = next
	}
}

// handleEnter updates the application state for the selected item on the active screen.
func (s *State) handleEnter() {
	switch s.nav.screen {
	case config.MenuScreen:
		s.activateMenuSelection()
	case config.LessonsScreen:
		s.activateLessonSelection()
	case config.SettingsScreen:
		s.activateSettingsSelection()
	case config.ResetConfirmationScreen:
		s.activateResetConfirmationSelection()
	default:
		s.nav.screen = config.MenuScreen
	}
}

func (s *State) activateMenuSelection() {
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
}

func (s *State) activateLessonSelection() {
	if len(s.lessons.available) == 0 {
		if s.lessons.err != nil {
			s.returnToMainMenu()
		}
		return
	}
	if s.nav.selection == len(s.lessons.available) {
		s.returnToMainMenu()
		return
	}

	s.lessons.activeIdx = s.nav.selection
	s.lessons.activePage = 0
	s.lessons.revealedHints = 0
	s.lessons.guide.GotoTop()
	s.nav.screen = config.LessonTerminalScreen
}

func (s *State) activateSettingsSelection() {
	if s.nav.selection == len(config.SettingsItems)-1 {
		s.returnToMainMenu()
		return
	}
	s.nav.screen = config.ResetConfirmationScreen
	s.nav.selection = 0
}

func (s *State) activateResetConfirmationSelection() {
	if s.nav.selection == 0 {
		s.nav.screen = config.SettingsScreen
		s.nav.selection = 0
	}
}

func (s *State) returnToMainMenu() {
	s.nav.screen = config.MenuScreen
	s.nav.selection = 0
}
