package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation moves the selection or acts on Enter for non-terminal
// screens. It returns tea.Quit when the learner quits, and otherwise the work
// started by entering another screen, if any.
func (s *State) handleNavigation(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, config.Keys.Quit):
		return tea.Quit
	case key.Matches(msg, config.Keys.Up):
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		if s.exitSelected() {
			return tea.Quit
		}
		return s.handleEnter()
	}
	return nil
}

// navigate is the only way to change screen. It shows to.screen with
// to.selection selected (the first item when left out) and returns the work
// that entering it starts, so no caller can change screen and forget that work.
func (s *State) navigate(nextState navigationState) tea.Cmd {
	// this is the old screen
	from := s.nav.screen
	// this is the new state
	s.nav = nextState
	return s.arrive(from)
}

// arrive starts the work the current screen needs when it is entered from
// another screen. Moving the selection never calls it, because only navigate
// changes screen.
func (s *State) arrive(from config.Screen) tea.Cmd {
	switch s.nav.screen {
	case config.LessonTerminalScreen:
		return s.startTerminalWithActiveLesson()
	case config.SandboxTerminalScreen:
		return s.startTerminal(nil)
	case config.ManageSandboxesScreen:
		// If we are currently in the manage sandboxes screen and we came from
		// the delete confirmation screen, do not do anything (keep selection)
		if from == config.DeleteSandboxesScreen {
			return nil
		}
		// else, start a new listing of sandboxes
		return s.settings.sandboxes.begin()
	default:
		return nil
	}
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
	case config.CursorStyleScreen:
		return len(ui.CursorStyles) + 1
	case config.ThemeScreen:
		return len(ui.Themes) + 2 // Auto, every theme, then Back
	case config.ResetConfirmationScreen:
		return len(config.ResetConfirmationItems)
	case config.ResetSettingsScreen:
		return len(config.ResetSettingsItems)
	case config.ManageSandboxesScreen:
		return s.manageSandboxesScreen().ItemCount()
	case config.DeleteSandboxesScreen:
		return 2 // No, go back; Yes, delete
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

// handleEnter acts on the selected item on the active screen and returns the
// work started by any screen it enters.
func (s *State) handleEnter() tea.Cmd {
	switch s.nav.screen {
	case config.MenuScreen:
		return s.activateMenuSelection()
	case config.LessonsScreen:
		return s.activateLessonSelection()
	case config.SettingsScreen:
		return s.activateSettingsSelection()
	case config.ResetConfirmationScreen:
		return s.activateResetConfirmationSelection()
	case config.ResetSettingsScreen:
		return s.returnToSettings(settingsResetSettingsIndex)
	case config.ManageSandboxesScreen:
		// Leaving stops the listing: its spinner halts and a late result is ignored.
		s.settings.sandboxes.isLoading = false
		return s.returnToSettings(settingsManageSandboxesIndex)
	default:
		return s.navigate(navigationState{screen: config.MenuScreen, selection: s.nav.selection})
	}
}
