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

// mainMenuBehaviour is the main menu's screen behaviour. Its handlers are inline
// because the menu is small.
func mainMenuBehaviour() screenBehaviour {
	return screenBehaviour{
		list: &listBehaviour{
			itemCount: func(State) int { return len(config.MenuItems) },
			// Enter opens the chosen screen or quits from Exit.
			onSelectItem: func(s *State) tea.Cmd {
				switch s.nav.selection {
				case menuSandboxIndex:
					return s.openScreen(navigationState{screen: config.SandboxTerminalScreen})
				case menuLessonsIndex:
					return s.openScreen(navigationState{screen: config.LessonsScreen})
				case menuSettingsIndex:
					// Opening Settings from the menu clears its last message.
					s.settings = settingsState{}
					return s.openScreen(navigationState{screen: config.SettingsScreen})
				case menuExitIndex:
					return tea.Quit
				default:
					return nil
				}
			},
		},
		view: func(s State, helpWidth int) string {
			return screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}
