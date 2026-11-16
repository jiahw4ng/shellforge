package app

import (
	"shellforge/internal/config"

	tea "charm.land/bubbletea/v2"
)

// screenBehaviour is everything that differs between screens. Shared handling
// (busy guards, help, list navigation, changeScreen) looks the active screen up
// with behaviour and calls these fields, so one entry describes one screen.
//
// A screen sets exactly one of onKeyPress and list.
//
// The fields follow the receiver rule in Update: onKeyPress and view take a
// State, while onSelectItem, onArrival, and onHelpToggle change the handler's
// copy through *State.
type screenBehaviour struct {
	// list is what plain list navigation needs. nil for a screen that sets
	// onKeyPress.
	list *listBehaviour
	// help says whether the help key works on the screen. Its zero value
	// means it does not.
	help helpBehaviour
	// onKeyPress handles every key on the screen. Only the terminal screens
	// set it; every other screen sets list instead.
	onKeyPress func(State, tea.KeyPressMsg) (tea.Model, tea.Cmd)
	// onArrival starts the work the screen needs when changeScreen enters it from
	// the screen given. nil means there is nothing to start.
	onArrival func(*State, config.Screen) tea.Cmd
	// view renders the screen; the second param is the width available for help.
	// Every screen has one.
	view func(State, int) string
}

// listBehaviour is what a list screen supplies to plain list navigation: Up
// and Down move the selection, Enter calls onSelectItem, and Quit quits (see
// the switch in handleKeyPress).
type listBehaviour struct {
	// itemCount is how many items Up and Down move between.
	itemCount func(State) int
	// onSelectItem acts on Enter on the selected item.
	onSelectItem func(*State) tea.Cmd
}

// helpBehaviour is how a screen handles the help key. The zero value turns it
// off, so on such a screen the key reaches onKeyPress like any other.
type helpBehaviour struct {
	// hasHelp is whether the help key toggles the full help on the screen.
	hasHelp bool
	// onHelpToggle is work the screen needs after help is toggled, such as
	// re-wrapping the lesson guide around the taller or shorter help. nil
	// means there is none.
	onHelpToggle func(*State)
}

// behaviour returns the active screen's behaviour. Screens with their own
// file build it with a function there, such as menuBehaviour in menu.go. A
// screen with no case gets an empty behaviour, whose nil view panics on the
// first render; TestEveryScreenHasABehaviour catches that before it ships.
func (s State) behaviour() screenBehaviour {
	switch s.nav.screen {
	case config.MenuScreen:
		return menuBehaviour()
	case config.SandboxTerminalScreen:
		return sandboxTerminalBehaviour()
	case config.LessonsScreen:
		return lessonsBehaviour()
	case config.LessonTerminalScreen:
		return lessonTerminalBehaviour()
	case config.SettingsScreen:
		return settingsBehaviour()
	case config.CursorStyleScreen:
		return cursorStyleBehaviour()
	case config.ThemeScreen:
		return themeBehaviour()
	case config.ResetConfirmationScreen:
		return resetLessonProgressBehaviour()
	case config.ResetSettingsScreen:
		return resetSettingsBehaviour()
	case config.ManageSandboxesScreen:
		return manageSandboxesBehaviour()
	case config.DeleteSandboxesScreen:
		return deleteSandboxesBehaviour()
	default:
		return screenBehaviour{}
	}
}
