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
		s.help.ShowAll = !s.help.ShowAll
		if s.nav.screen == config.LessonTerminalScreen {
			s.prepareLessonGuide()
		}
		return s, nil
	}

	// if the current screen defined special behaviour for key pressess,
	// perform that behvaiour
	if keyPress := s.behaviour().onKeyPress; keyPress != nil {
		return keyPress(s, msg)
	}

	// else, handle basic list navigation (up/down, enter, quit)
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
