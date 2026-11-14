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

// handleNavigationKeyPress updates menu selection or screen state, quits when
// requested, and schedules terminal startup after entering a terminal screen
// or a sandbox listing after entering Manage sandboxes.
func (s State) handleNavigationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	previous := s.nav.screen
	// handle ordinary navigation and quit requests, returning early if the app should quit
	if shouldQuit := s.handleNavigation(msg); shouldQuit {
		return s, tea.Quit
	}

	// after navigating, check if the user entered a screen that
	// requires starting a terminal session or loading data
	switch s.nav.screen {
	case config.LessonTerminalScreen:
		// start terminal with selected lesson
		return s, s.startTerminalWithActiveLesson()
	case config.SandboxTerminalScreen:
		// start terminal with no lesson
		return s, s.startTerminal(nil)
	case config.ManageSandboxesScreen:
		// load the sandboxes for the listing
		// why is previous necessary? TODO
		if previous == config.ManageSandboxesScreen {
			return s, nil
		}
		return s, s.settings.sandboxes.begin()
	default:
		return s, nil
	}
}

func (s *State) toggleHelp() {
	s.help.ShowAll = !s.help.ShowAll
	if s.nav.screen == config.LessonTerminalScreen {
		s.prepareLessonGuide()
	}
}
