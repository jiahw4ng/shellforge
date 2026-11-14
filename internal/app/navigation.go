package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation updates non-terminal navigation and reports whether the app should quit.
func (s *State) handleNavigation(msg tea.KeyMsg) bool {
	switch {
	case key.Matches(msg, config.Keys.Quit):
		return true
	case key.Matches(msg, config.Keys.Up):
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		if s.exitSelected() {
			return true
		}
		s.handleEnter()
	}

	return false
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

// handleEnter updates the application state for the selected item on the active screen.
func (s *State) handleEnter() {
	switch s.nav.screen {
	case config.MenuScreen:
		s.activateMenuSelection()
	case config.LessonsScreen:
		s.activateLessonSelection()
	case config.SettingsScreen:
		s.activateSettingsSelection()
	case config.ResetConfirmationScreen:
		s.activateResetConfirmationSelection()
	case config.ResetSettingsScreen:
		s.returnToSettings(settingsResetSettingsIndex)
	case config.ManageSandboxesScreen:
		// Leaving stops the listing: its spinner halts and a late result is ignored.
		s.settings.sandboxes.isLoading = false
		s.returnToSettings(settingsManageSandboxesIndex)
	default:
		s.nav.screen = config.MenuScreen
	}
}
