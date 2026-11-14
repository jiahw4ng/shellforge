package app

import "shellforge/internal/config"

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

func (s *State) returnToMainMenu() {
	s.nav.screen = config.MenuScreen
	s.nav.selection = 0
}

func (s State) exitSelected() bool {
	return s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1
}
