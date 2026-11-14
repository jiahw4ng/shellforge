package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// screenBehaviour is everything that differs between screens. Shared handling
// (busy guards, help, list navigation, navigate) looks the active screen up
// with behaviour and calls these fields, so one entry describes one screen.
//
// The fields follow the receiver rule in Update: keyPress and view take a
// State, while selectItem and arrive change the handler's copy through *State.
type screenBehaviour struct {
	// keyPress handles every key on the screen. nil means plain list
	// navigation: Up and Down move the selection, Enter calls selectItem, and
	// Quit quits.
	keyPress func(State, tea.KeyPressMsg) (tea.Model, tea.Cmd)
	// itemCount is how many items Up and Down move between. nil means none.
	itemCount func(State) int
	// selectItem acts on Enter when list navigation handles it. nil means
	// Enter does nothing.
	selectItem func(*State) tea.Cmd
	// arrive starts the work the screen needs when navigate enters it from
	// the screen given. nil means there is nothing to start.
	arrive func(*State, config.Screen) tea.Cmd
	// view renders the screen; the int is the width available for help.
	// Every screen has one.
	view func(State, int) string
}

// behaviour returns the active screen's behaviour. A screen missing from
// behaviourOf behaves as the main menu, as every per-screen switch's default
// did; TestEveryScreenHasABehaviour keeps any screen from relying on that.
func (s State) behaviour() screenBehaviour {
	if behaviour, ok := behaviourOf(s.nav.screen); ok {
		return behaviour
	}
	behaviour, _ := behaviourOf(config.MenuScreen)
	return behaviour
}

// behaviourOf returns the behaviour of screen, and false for a screen that
// has no entry.
func behaviourOf(screen config.Screen) (behaviour screenBehaviour, ok bool) {
	switch screen {
	case config.MenuScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(config.MenuItems) },
			selectItem: (*State).activateMenuSelection,
			view:       State.menuView,
		}, true
	case config.SandboxTerminalScreen:
		return screenBehaviour{
			keyPress: State.handleTerminalKeyPress,
			arrive:   func(s *State, _ config.Screen) tea.Cmd { return s.startTerminal(nil) },
			view:     func(s State, _ int) string { return s.sandboxView() },
		}, true
	case config.LessonsScreen:
		return screenBehaviour{
			itemCount:  State.lessonsItemCount,
			selectItem: (*State).activateLessonSelection,
			view:       State.lessonsView,
		}, true
	case config.LessonTerminalScreen:
		return screenBehaviour{
			keyPress: State.handleLessonKeyPress,
			arrive:   func(s *State, _ config.Screen) tea.Cmd { return s.startTerminalWithActiveLesson() },
			view:     func(s State, _ int) string { return s.lessonView() },
		}, true
	case config.SettingsScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(config.SettingsItems) },
			selectItem: (*State).activateSettingsSelection,
			view:       State.settingsView,
		}, true
	case config.CursorStyleScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(ui.CursorStyles) + 1 }, // every style, then Back
			selectItem: (*State).activateCursorStyleSelection,
			view:       State.cursorStyleView,
		}, true
	case config.ThemeScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(ui.Themes) + 2 }, // Auto, every theme, then Back
			selectItem: (*State).activateThemeSelection,
			view:       State.themeView,
		}, true
	case config.ResetConfirmationScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(config.ResetConfirmationItems) },
			selectItem: (*State).activateResetConfirmationSelection,
			view:       State.resetConfirmationView,
		}, true
	case config.ResetSettingsScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return len(config.ResetSettingsItems) },
			selectItem: (*State).activateResetSettingsSelection,
			view:       State.resetSettingsView,
		}, true
	case config.ManageSandboxesScreen:
		return screenBehaviour{
			itemCount:  State.manageSandboxesItemCount,
			selectItem: (*State).activateManageSandboxesSelection,
			arrive:     (*State).arriveAtManageSandboxes,
			view:       State.manageSandboxesView,
		}, true
	case config.DeleteSandboxesScreen:
		return screenBehaviour{
			itemCount:  func(State) int { return 2 }, // No, go back; Yes, delete
			selectItem: (*State).activateDeleteSandboxesSelection,
			view:       State.deleteSandboxesView,
		}, true
	default:
		return screenBehaviour{}, false
	}
}
