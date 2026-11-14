package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress handles only Ctrl+C while blocking work runs, and otherwise
// delegates the keypress to the handler for the active screen.
func (s State) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.blockingWork() {
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

	if keyPress := s.behaviour().keyPress; keyPress != nil {
		return keyPress(s, msg)
	}
	return s.handleNavigationKeyPress(msg)
}

// blockingWork reports that a reset or a sandbox deletion is running. Every
// key except Ctrl+C is ignored meanwhile, so Enter cannot start the work again
// and the learner cannot leave the screen that reports its result.
//
// Quitting is always allowed. A reset is one SQLite statement, so it either
// happened or did not. A deletion removes sandboxes one at a time, so quitting
// part-way can leave some of them; Manage sandboxes lists whatever remains.
func (s State) blockingWork() bool {
	return s.settings.isResetting || s.sandboxes.isDeleting
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
