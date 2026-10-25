package app

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update handles Bubble Tea events and advances the application state.
// Note: returning nil for the second value means “no work to schedule.”
// A tea.Cmd is deferred work—a function Bubble Tea runs that
// eventually returns another message.
// it has the type func() tea.Msg, so it can be executed later and return a message to Update.
// All the functions invoked by Update must have the same return signature
// for easy delegation. They can return nil for the second value if they have no work to schedule.
// as long as a tea.Cmd is propagated up the call stack to this method,
// bubble tea can exexcute it and send the resulting message back to Update.
// note: bubble tea manages scheduling and execution, so command A sent after B may be processed before B, and the order of messages is not guaranteed.
// also, a tea.Cmd can return nil if no further message is needed, so Update won't be called once the command completes.
// TLDR: Update() handles an event, updates the model, and optionally
// schedules more work using a command. That work may produce another event,
// which calls Update() and continues the cycle.
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
