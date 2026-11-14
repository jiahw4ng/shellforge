package app

import (
	"shellforge/internal/config"

	tea "charm.land/bubbletea/v2"
)

func (s *State) activateMenuSelection() tea.Cmd {
	switch s.nav.selection {
	case 0:
		return s.navigate(config.SandboxTerminalScreen, 0)
	case 1:
		return s.navigate(config.LessonsScreen, 0)
	case 2:
		s.settings = settingsState{}
		return s.navigate(config.SettingsScreen, 0)
	default:
		return nil
	}
}

func (s *State) returnToMainMenu() tea.Cmd {
	return s.navigate(config.MenuScreen, 0)
}

func (s State) exitSelected() bool {
	return s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1
}
