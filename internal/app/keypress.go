package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress leaves state unchanged while a reset is running and otherwise
// delegates the keypress to the handler for the active screen.
func (s State) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.settings.isResetting {
		return s, nil
	}
	if s.settings.sandboxes.isDeleting {
		if key.Matches(msg, config.Keys.Quit) {
			return s, tea.Quit
		}
		return s, nil
	}

	// toggle help on any screen except the sandbox terminal, which doesnt have any help to show
	if s.nav.screen != config.SandboxTerminalScreen && key.Matches(msg, config.Keys.ToggleHelp) {
		s.toggleHelp()
		return s, nil
	}

	switch s.nav.screen {
	case config.LessonTerminalScreen:
		return s.handleLessonKeyPress(msg)
	case config.SandboxTerminalScreen:
		return s.handleTerminalKeyPress(msg)
	case config.ResetConfirmationScreen:
		return s.handleResetConfirmationKeyPress(msg)
	case config.ResetSettingsScreen:
		return s.handleResetSettingsKeyPress(msg)
	case config.CursorStyleScreen:
		return s.handleCursorStyleKeyPress(msg)
	case config.ThemeScreen:
		return s.handleThemeKeyPress(msg)
	case config.ManageSandboxesScreen:
		return s.handleManageSandboxesKeyPress(msg)
	case config.DeleteSandboxesScreen:
		return s.handleDeleteSandboxesKeyPress(msg)
	default:
		return s.handleNavigationKeyPress(msg)
	}
}

// handleNavigationKeyPress moves the selection, acts on Enter, or quits.
// Entering a screen starts its work through navigate.
func (s State) handleNavigationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	command := s.handleNavigation(msg)
	return s, command
}

func (s *State) toggleHelp() {
	s.help.ShowAll = !s.help.ShowAll
	if s.nav.screen == config.LessonTerminalScreen {
		s.prepareLessonGuide()
	}
}
