package app

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

const logErrorKey = "error"

// handleWindowSize stores the latest viewport dimensions, rebuilds an active
// lesson guide, and schedules a resize for an active terminal session.
func (s State) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	s.viewport.width = msg.Width
	s.viewport.height = msg.Height
	s.prepareLessonGuide()
	if s.term.session != nil {
		return s, s.resizeTerminal()
	}
	return s, nil
}

// handleSpinnerTick advances the terminal startup spinner while a session is
// starting and the sandbox listing spinner while a listing runs. Each spinner
// ignores ticks addressed to the other, so a tick reaches only its owner.
func (s State) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd
	if s.term.isStarting {
		updated, command := s.term.spinner.Update(msg)
		s.term.spinner = updated
		commands = append(commands, command)
	}
	if s.settings.sandboxes.isLoading || s.settings.sandboxes.isDeleting {
		updated, command := s.settings.sandboxes.spinner.Update(msg)
		s.settings.sandboxes.spinner = updated
		commands = append(commands, command)
	}
	return s, tea.Batch(commands...)
}

// handleOther forwards non-key events to an active terminal session and
// otherwise leaves the application state unchanged.
func (s State) handleOther(message tea.Msg) (tea.Model, tea.Cmd) {
	if s.usesTerminal() && s.term.session != nil {
		return s, s.term.session.Update(message)
	}
	return s, nil
}
