package app

import (
	"shellforge/internal/config"

	tea "charm.land/bubbletea/v2"
)

// screenBehaviour is everything that differs between screens. Shared handling
// (busy guards, help, list navigation, openScreen and goBack) looks the active screen up
// with activeScreenBehaviour and calls these fields, so one entry describes
// one screen.
//
// A screen sets exactly one of onKeyPress and list.
//
// The fields follow the receiver rule in Update: onKeyPress and view take a
// State, while onSelectItem, onArrival, onDeparture, and onHelpToggle change
// the handler's copy through *State.
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
	// onArrival starts the work the screen needs when openScreen opens it
	// from the screen given, its parent. Returning from a child does not run
	// it. nil means there is nothing to start.
	onArrival func(*State, config.Screen) tea.Cmd
	// onDeparture does the work the screen needs when goBack closes it for
	// the screen given, its parent. It runs while the screen is still active.
	// Opening a child does not run it. nil means there is nothing to do.
	onDeparture func(*State, config.Screen) tea.Cmd
	// view renders the screen; the second param is the width available for help.
	// Every screen has one.
	view func(State, int) string
	// parent is the screen this one is opened from and goBack returns to, so
	// the parents form a tree rooted at the main menu. The parent stays open
	// while this screen is: a deletion confirmation keeps Manage sandboxes'
	// listing and ticks. nil only for the main menu; it is a pointer because
	// the zero config.Screen is the menu.
	parent *config.Screen
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

// activeScreenBehaviour returns the active screen's behaviour.
func (s State) activeScreenBehaviour() screenBehaviour {
	return behaviourOf(s.nav.screen)
}

// behaviourOf returns a specific screen's behaviour. Screens with their own
// file build it with a function there, such as menuBehaviour in menu.go. A
// screen with no case gets an empty behaviour, whose nil view panics on the
// first render; TestEveryScreenHasABehaviour catches that before it ships.
func behaviourOf(screen config.Screen) screenBehaviour {
	switch screen {
	case config.MainMenuScreen:
		return mainMenuBehaviour()
	case config.SandboxTerminalScreen:
		return sandboxTerminalBehaviour()
	case config.LessonsScreen:
		return lessonsBehaviour()
	case config.LessonTerminalScreen:
		return lessonTerminalBehaviour()
	case config.SettingsScreen:
		return settingsBehaviour()
	case config.ChooseCursorStyleScreen:
		return cursorStyleBehaviour()
	case config.ChooseThemeScreen:
		return themeBehaviour()
	case config.ResetLessonProgressScreen:
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
