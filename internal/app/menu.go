package app

import (
	"shellforge/internal/config"
	"shellforge/internal/screens"

	tea "charm.land/bubbletea/v2"
)

// activateMenuSelection opens the sandbox, the lesson list, or Settings, or
// quits from Exit, the last item.
func (s *State) activateMenuSelection() tea.Cmd {
	switch s.nav.selection {
	case 0:
		return s.navigate(navigationState{screen: config.SandboxTerminalScreen})
	case 1:
		return s.navigate(navigationState{screen: config.LessonsScreen})
	case 2:
		s.settings = settingsState{}
		return s.navigate(navigationState{screen: config.SettingsScreen})
	case len(config.MenuItems) - 1:
		return tea.Quit
	default:
		return nil
	}
}

func (s *State) returnToMainMenu() tea.Cmd {
	return s.navigate(navigationState{screen: config.MenuScreen})
}

// menuView shows the main menu.
func (s State) menuView(helpWidth int) string {
	return screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
}
