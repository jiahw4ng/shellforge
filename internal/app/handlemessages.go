package app

import (
	"errors"
	"log/slog"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/samber/lo"
)

const (
	logErrorKey = "error"
	// lightThemeID is the theme detected on a light terminal background.
	lightThemeID = "light"
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
		slog.Error("Could not load lesson completion", logErrorKey, msg.err)
		s.lessons.progressNotLoaded = true
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

// handleTerminalStartProgress records the next startup step of the current
// attempt and waits for the one after it.
func (s State) handleTerminalStartProgress(msg termStartProgressMsg) (tea.Model, tea.Cmd) {
	// Steps from an older attempt are ignored.
	if msg.gen != s.term.gen || !s.term.isStarting {
		return s, nil
	}
	// append the new step to the list of steps so far, and wait for the next one
	s.term.startSteps = append(s.term.startSteps, msg.step)
	return s, waitForStartEvent(s.term.startEvents)
}

// handleTerminalStarted ignores stale results, records startup failures, or
// installs the new session and resets terminal-scoped lesson progress.
func (s State) handleTerminalStarted(msg termStartedMsg) (tea.Model, tea.Cmd) {
	// Steps from an older attempt are ignored.
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	// clear the starting steps
	s.term.startSteps = nil
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
	return s, tea.Batch(s.term.session.Init(), waitForTerminalExit(s.term.session.Exited, s.term.gen), s.resizeTerminal())
}

// handleTerminalExited ignores stale exits, stops in-flight terminal work, and
// either returns from an intentional exit or records an unexpected failure.
func (s State) handleTerminalExited(msg termExitedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.term.isStarting = false
	s.lessons.progress.isChecking = false
	if s.term.hasRequestedExit {
		s.finishRequestedExit()
		return s, nil
	}
	s.recordUnexpectedExit()
	return s, nil
}

// finishRequestedExit closes the terminal and returns to the screen that
// launched it, leaving no failure behind.
func (s *State) finishRequestedExit() {
	s.CloseTerminal()
	s.term.clearFailure()
	s.returnFromTerminal()
}

// recordUnexpectedExit keeps the terminal's last frame for the learner, closes
// it, and records the failure. The frame is read first: closing drops the session.
func (s *State) recordUnexpectedExit() {
	s.term.output = s.term.session.View()
	s.CloseTerminal()
	s.term.err = errors.New("the terminal closed unexpectedly")
}

// handleAssertionsChecked stores current-generation assertion results and marks
// a newly passing lesson complete, scheduling persistence when available.
func (s State) handleAssertionsChecked(msg assertionsCheckedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.recordAssertionResults(msg.results)

	hasPassedAllAssertions := lo.EveryBy(msg.results, func(result assertion.Result) bool {
		return result.Passed
	})
	if msg.lessonID == "" || !hasPassedAllAssertions {
		return s, nil
	}
	return s, s.completeLesson(msg.lessonID)
}

// recordAssertionResults stores the latest check outcome and shows it.
func (s *State) recordAssertionResults(results []assertion.Result) {
	s.lessons.progress.isChecking = false
	s.lessons.progress.hasChecked = true
	s.lessons.progress.results = results
	s.lessons.progress.assertionsExpanded = true
}

// completeLesson marks a lesson complete the first time it passes. It returns
// the command that persists the completion, or nil when the lesson was already
// complete or storage is unavailable.
func (s *State) completeLesson(lessonID string) tea.Cmd {
	if s.lessons.completed[lessonID] {
		return nil
	}
	s.lessons.completed[lessonID] = true
	if s.lessons.store == nil {
		return nil
	}
	return saveCompletion(s.lessons.store, lessonID)
}

// handleCompletionSaved logs any persistence error and flags it for the help
// notice. The lesson was already marked complete in memory, so completion
// itself is unchanged.
func (s State) handleCompletionSaved(msg completionSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		slog.Error("Could not persist lesson completion", "lesson_id", msg.lessonID, logErrorKey, msg.err)
		s.lessons.progressNotSaved = true
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
		slog.Error("Could not reset lesson completion", logErrorKey, msg.err)
		return s, nil
	}
	s.lessons.completed = make(map[string]bool)
	s.lessons.progress = progressState{}
	// nothing is left to lose or to read back, so the storage notice no longer applies
	s.lessons.progressNotLoaded = false
	s.lessons.progressNotSaved = false
	s.settings.message = "Lesson progress has been reset."
	s.settings.failed = false
	return s, nil
}

// handleSettingsReset reports a storage failure in Settings. The defaults stay
// in effect for this session, but the old saved values return on restart.
func (s State) handleSettingsReset(msg settingsResetMsg) (tea.Model, tea.Cmd) {
	if msg.err == nil {
		return s, nil
	}
	slog.Error("Could not reset settings", logErrorKey, msg.err)
	s.settings.message = "Settings could not be reset in storage; the old values return when Shellforge restarts."
	s.settings.failed = true
	return s, nil
}

// handleCursorStyleSaved reports a persistence failure in Settings while keeping
// the chosen cursor style for the rest of the session.
func (s State) handleCursorStyleSaved(msg cursorStyleSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err == nil {
		return s, nil
	}
	slog.Error("Could not persist cursor style", "cursor_style", msg.styleID, logErrorKey, msg.err)
	s.settings.message = "Cursor style could not be saved; it applies until Shellforge exits."
	s.settings.failed = true
	return s, nil
}

// handleThemeSaved reports a persistence failure in Settings while keeping the
// chosen theme for the rest of the session.
func (s State) handleThemeSaved(msg themeSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err == nil {
		return s, nil
	}
	slog.Error("Could not persist theme", "theme", msg.themeID, logErrorKey, msg.err)
	s.settings.message = "Theme could not be saved; it applies until Shellforge exits."
	s.settings.failed = true
	return s, nil
}

// handleBackgroundColor applies Light or Dark from the terminal's reported
// background while the theme is in Auto mode. The detected theme is not
// persisted; only the "auto" mode is, and only once, when nothing says so yet.
// A reply that arrives after the learner picked a theme is ignored.
func (s State) handleBackgroundColor(msg tea.BackgroundColorMsg) (tea.Model, tea.Cmd) {
	if s.preferences.themeExplicit {
		return s, nil
	}
	slog.Debug("Detected terminal background color", "dark", msg.IsDark())
	themeID := ui.DefaultTheme.ID
	if !msg.IsDark() {
		themeID = lightThemeID
	}
	theme, ok := ui.ThemeByID(themeID)
	if !ok {
		return s, nil
	}
	s.setTheme(theme)
	if s.preferences.store == nil || s.preferences.autoStored {
		return s, nil
	}
	s.preferences.autoStored = true
	return s, saveTheme(s.preferences.store, ui.AutoThemeID)
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

// handleOther forwards non-key events to an active terminal session and
// otherwise leaves the application state unchanged.
func (s State) handleOther(message tea.Msg) (tea.Model, tea.Cmd) {
	if s.usesTerminal() && s.term.session != nil {
		return s, s.term.session.Update(message)
	}
	return s, nil
}
