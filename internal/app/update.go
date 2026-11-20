package app

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update handles one Bubble Tea message and returns the new state plus any
// work to schedule. That work may produce another message, which calls Update
// again and continues the cycle.
//
// Commands: a tea.Cmd is deferred work, a func() tea.Msg that Bubble Tea runs
// in its own goroutine and whose result it sends back to Update. A nil tea.Cmd
// means there is no work to schedule, and a command that returns a nil
// message ends without calling Update again. A command only runs if it is
// returned all the way up HERE (Update), so every helper on the way must pass its
// tea.Cmd back to its caller. Commands run concurrently, so their messages can
// arrive in any order; use tea.Sequence when order matters.
//
// Receivers (architecture.md, rule 2): handlers that return the new model,
// such as func (s State) handleXYZ(msg) (tea.Model, tea.Cmd), use VALUE
// receivers. Each gets its own copy of State to change and return. Helpers that
// only change state or return a tea.Cmd, such as func (s *State)
// openChildScreen(...) tea.Cmd, use POINTER receivers, so they change the handler's
// copy in place. The smaller state types follow the same pattern, for example
// sandboxListState's begin and chosenSandboxes.
//
// The copy is SHALLOW: maps (lessons.completed, sandboxes.chosen) and
// pointers (term.session) are shared with the previous state. That is safe
// because Bubble Tea always replaces the old model with the returned one; never
// keep an old State and expect those parts of it to stay unchanged.
func (s State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case nil:
		// nil probably will never be an argument to Update, since
		// a command that returns tea.Msg = nil will be treated as
		// "no message to send", so Update wont even be called.
		// but if it is, just ignore it and return the current state.
		return s, nil
	case lessonsLoadedMsg:
		return s.handleLessonsLoaded(msg)
	case completionsLoadedMsg:
		return s.handleCompletionsLoaded(msg)
	case tea.WindowSizeMsg:
		return s.handleWindowSize(msg)
	case termStartProgressMsg:
		return s.handleTerminalStartProgress(msg)
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
	case settingsResetMsg:
		return s.handleSettingsReset(msg)
	case cursorStyleSavedMsg:
		return s.handleCursorStyleSaved(msg)
	case themeSavedMsg:
		return s.handleThemeSaved(msg)
	case sandboxesListedMsg:
		return s.handleSandboxesListed(msg)
	case sandboxesDeletedMsg:
		return s.handleSandboxesDeleted(msg)
	case tea.BackgroundColorMsg:
		return s.handleBackgroundColor(msg)
	case spinner.TickMsg:
		return s.handleSpinnerTick(msg)
	case tea.KeyPressMsg:
		return s.handleKeyPress(msg)
	default:
		return s.handleOther(message)
	}
}
