package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleKeyPress handles only Ctrl+C while blocking work runs. Otherwise it
// toggles help on a screen that has it, then passes the key to the screen's
// onKeyPress, or does plain list navigation for a list screen.
func (s State) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if s.blockingWork() {
		if key.Matches(msg, config.Keys.Quit) {
			return s, tea.Quit
		}
		return s, nil
	}

	behv := s.activeScreenBehaviour()

	// Toggle help on any screen that has it. The standalone sandbox does not,
	// so there the key goes to the shell.
	if behv.help.hasHelp && key.Matches(msg, config.Keys.ToggleHelp) {
		s.help.ShowAll = !s.help.ShowAll

		// &s, so the hook changes this handler's copy, which is returned below.
		if behv.help.onHelpToggle != nil {
			behv.help.onHelpToggle(&s)
		}

		return s, nil
	}

	// A screen with its own key handling (the terminals) gets every other key.
	if behv.onKeyPress != nil {
		return behv.onKeyPress(s, msg)
	}

	// Every other screen is a list: Quit, Up, Down, and Enter.
	var commandToDo tea.Cmd

	switch {
	case key.Matches(msg, config.Keys.Quit):
		commandToDo = tea.Quit
	case key.Matches(msg, config.Keys.Up):
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		commandToDo = s.handleEnter()
	}

	return s, commandToDo
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
