package app

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea events and advances the application state.
func (s State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case lessonsLoadedMsg:
		return s.handleLessonsLoaded(msg)
	case completionsLoadedMsg:
		return s.handleCompletionsLoaded(msg)
	case tea.WindowSizeMsg:
		return s.handleWindowSize(msg)
	case termStartedMsg:
		return s.handleTerminalStarted(msg)
	case termExitedMsg:
		return s.handleTerminalExited(msg)
	case assertionsCheckedMsg:
		return s.handleAssertionsChecked(msg)
	case completionSavedMsg:
		return s.handleCompletionSaved(msg)
	case completionsResetMsg:
		return s.handleCompletionsReset(msg)
	case spinner.TickMsg:
		return s.handleSpinnerTick(msg)
	case tea.KeyPressMsg:
		return s.handleKeyPress(msg)
	default:
		return s.handleOtherMessage(message)
	}
}
