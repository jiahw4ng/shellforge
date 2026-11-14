package app

import (
	"shellforge/internal/config"
	"shellforge/internal/screens"

	tea "charm.land/bubbletea/v2"
)

func (s *State) activateMenuSelection() tea.Cmd {
	switch s.nav.selection {
	case 0:
		return s.navigate(navigationState{screen: config.SandboxTerminalScreen})
	case 1:
		return s.navigate(navigationState{screen: config.LessonsScreen})
	case 2:
		s.settings = settingsState{}
		return s.navigate(navigationState{screen: config.SettingsScreen})
	default:
		return nil
	}
}

func (s *State) returnToMainMenu() tea.Cmd {
	return s.navigate(navigationState{screen: config.MenuScreen})
}

func (s State) exitSelected() bool {
	return s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1
}

// menuView shows the main menu.
func (s State) menuView(helpWidth int) string {
	return screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
}
