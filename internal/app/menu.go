package app

import (
	"shellforge/internal/config"
	"shellforge/internal/screens"

	tea "charm.land/bubbletea/v2"
)

// Main menu item positions, in the order of config.MenuItems.
const (
	menuSandboxIndex = iota
	menuLessonsIndex
	menuSettingsIndex
	menuExitIndex
)

// activateMenuSelection opens the sandbox, the lesson list, or Settings, or
// quits from Exit.
func (s *State) activateMenuSelection() tea.Cmd {
	switch s.nav.selection {
	case menuSandboxIndex:
		return s.navigate(navigationState{screen: config.SandboxTerminalScreen})
	case menuLessonsIndex:
		return s.navigate(navigationState{screen: config.LessonsScreen})
	case menuSettingsIndex:
		s.settings = settingsState{}
		return s.navigate(navigationState{screen: config.SettingsScreen})
	case menuExitIndex:
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
