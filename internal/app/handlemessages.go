package app

import (
	"errors"
	"log/slog"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/terminal"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/samber/lo"
)

// handleLessonsLoaded records the load result and replaces the available
// lessons only when loading succeeded.
func (s State) handleLessonsLoaded(msg lessonsLoadedMsg) (tea.Model, tea.Cmd) {
	s.lessons.err = msg.err
	if msg.err == nil {
		s.lessons.available = msg.lessons
	}
	return s, nil
}

// handleCompletionsLoaded merges persisted lesson IDs into the completion map,
// leaving the current completion state unchanged when loading failed.
func (s State) handleCompletionsLoaded(msg completionsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		slog.Error("Could not load lesson completion", "error", msg.err)
		return s, nil
	}
	for _, lessonID := range msg.lessonIDs {
		s.lessons.completed[lessonID] = true
	}
	return s, nil
}

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

// handleTerminalStarted ignores stale results, records startup failures, or
// installs the new session and resets terminal-scoped lesson progress.
func (s State) handleTerminalStarted(msg termStartedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	if msg.err != nil {
		s.term.err = msg.err
		return s, nil
	}
	if msg.session == nil {
		s.term.err = terminal.NewInvalidStartResultError()
		return s, nil
	}
	s.term.session = msg.session
	s.term.output = ""
	s.term.hasRequestedExit = false
	s.lessons.progress = progressState{}
	return s, tea.Batch(s.term.session.Init(), waitForTerminalExit(s.term.session.Exited(), s.term.gen), s.resizeTerminal())
}

// handleTerminalExited ignores stale exits, stops in-flight terminal work, and
// either returns from an intentional exit or records an unexpected failure.
func (s State) handleTerminalExited(msg termExitedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	s.lessons.progress.isChecking = false
	if !s.term.hasRequestedExit {
		s.term.output = s.term.session.View()
	}
	s.CloseTerminal()
	if s.term.hasRequestedExit {
		s.term.output = ""
		s.term.err = nil
		s.returnFromTerminal()
	} else {
		s.term.err = errors.New("the terminal closed unexpectedly")
	}
	return s, nil
}

// handleAssertionsChecked stores current-generation assertion results and marks
// a newly passing lesson complete, scheduling persistence when available.
func (s State) handleAssertionsChecked(msg assertionsCheckedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.lessons.progress.isChecking = false
	s.lessons.progress.hasChecked = true
	s.lessons.progress.results = msg.results

	hasPassedAllAssertions := lo.EveryBy(msg.results, func(result assertion.Result) bool {
		return result.Passed
	})

	if msg.lessonID == "" || !hasPassedAllAssertions {
		return s, nil
	}
	if s.lessons.completed[msg.lessonID] {
		return s, nil
	}
	s.lessons.completed[msg.lessonID] = true
	if s.lessons.store != nil {
		return s, saveCompletion(s.lessons.store, msg.lessonID)
	}
	return s, nil
}

// handleCompletionSaved leaves state unchanged and logs any persistence error
// because the lesson was already marked complete in memory.
func (s State) handleCompletionSaved(msg completionSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		slog.Error("Could not persist lesson completion", "lesson_id", msg.lessonID, "error", msg.err)
	}
	return s, nil
}

// handleCompletionsReset returns to Settings and records either the reset
// failure or a cleared completion and progress state.
func (s State) handleCompletionsReset(msg completionsResetMsg) (tea.Model, tea.Cmd) {
	s.settings.isResetting = false
	s.nav.screen = config.SettingsScreen
	s.nav.selection = 0
	if msg.err != nil {
		s.settings.message = "Lesson progress could not be reset."
		s.settings.failed = true
		slog.Error("Could not reset lesson completion", "error", msg.err)
		return s, nil
	}
	s.lessons.completed = make(map[string]bool)
	s.lessons.progress = progressState{}
	s.settings.message = "Lesson progress has been reset."
	s.settings.failed = false
	return s, nil
}

// handleSpinnerTick advances the terminal startup spinner only while a session
// is starting, leaving all state unchanged after startup finishes.
func (s State) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if !s.term.isStarting {
		return s, nil
	}
	updated, command := s.term.spinner.Update(msg)
	s.term.spinner = updated
	return s, command
}

// handleOtherMessage forwards non-key events to an active terminal session and
// otherwise leaves the application state unchanged.
func (s State) handleOtherMessage(message tea.Msg) (tea.Model, tea.Cmd) {
	if s.usesTerminal() && s.term.session != nil {
		return s, s.term.session.Update(message)
	}
	return s, nil
}
