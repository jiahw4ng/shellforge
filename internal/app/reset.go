package app

import (
	"context"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/screens"
	"shellforge/internal/storage"
	"shellforge/internal/ui"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type completionsResetMsg struct {
	err error
}

type settingsResetMsg struct {
	err error
}

func resetLessonCompletions(store storage.CompletionStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.ResetLessonCompletions(ctx)
		return completionsResetMsg{err: err}
	}
}

func resetStoredSettings(store storage.SettingsStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return settingsResetMsg{err: store.ResetSettings(ctx)}
	}
}

// reportResetUnavailable leaves the reset confirmation for Settings and tells
// the learner that nothing can be reset without storage.
func (s *State) reportResetUnavailable() tea.Cmd {
	back := s.returnToSettings(0)
	s.settings.message = "Lesson progress could not be reset because storage is unavailable."
	s.settings.failed = true
	return back
}

// handleResetSettingsKeyPress delegates ordinary navigation and, when the
// confirming choice is selected, restores the default settings.
func (s State) handleResetSettingsKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if !key.Matches(msg, config.Keys.SelectItem) || s.nav.selection != 1 {
		return s.handleNavigationKeyPress(msg)
	}
	return s, s.resetSettings()
}

// handleResetConfirmationKeyPress delegates ordinary navigation, reports unavailable
// storage, or marks reset as active and schedules completion deletion.
func (s State) handleResetConfirmationKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// if the user did not press Enter on the correct option, handle navigation normally
	if !key.Matches(msg, config.Keys.SelectItem) || s.nav.selection != 1 {
		return s.handleNavigationKeyPress(msg)
	}

	// guard againt unavailable storage
	if s.lessons.store == nil {
		back := s.reportResetUnavailable()
		return s, back
	}

	s.settings.isResetting = true
	return s, resetLessonCompletions(s.lessons.store)
}

// resetSettings restores the default cursor style and Auto theme, returns to
// Settings, and clears the saved settings when storage is available. Auto
// starts from the default theme and asks the terminal for its background
// again, exactly as on a first launch. Lesson completions are left alone.
func (s *State) resetSettings() tea.Cmd {
	s.preferences.cursorStyleID = ""
	s.preferences.themeExplicit = false
	s.preferences.autoStored = false
	s.setTheme(ui.DefaultTheme)
	back := s.returnToSettings(settingsResetSettingsIndex)

	if s.preferences.store == nil {
		s.settings.message = "Settings were reset until Shellforge exits; settings storage is unavailable."
		s.settings.failed = true
		return tea.Batch(back, tea.RequestBackgroundColor)
	}
	s.settings.message = "Settings have been reset."
	s.settings.failed = false
	return tea.Batch(back, resetStoredSettings(s.preferences.store), tea.RequestBackgroundColor)
}

func (s *State) activateResetConfirmationSelection() tea.Cmd {
	if s.nav.selection == 0 {
		return s.returnToSettings(0)
	}
	return nil
}

// resetConfirmationView hides the help line while a reset is in progress.
func (s State) resetConfirmationView(helpWidth int) string {
	help := ""
	if !s.settings.isResetting {
		help = s.navigationHelp(helpWidth)
	}
	return screens.ResetConfirmation(config.ResetConfirmationItems, s.nav.selection, s.settings.isResetting, help)
}

// handleCompletionsReset returns to Settings and records either the reset
// failure or a cleared completion and progress state.
func (s State) handleCompletionsReset(msg completionsResetMsg) (tea.Model, tea.Cmd) {
	s.settings.isResetting = false
	back := s.returnToSettings(0)
	if msg.err != nil {
		s.settings.message = "Lesson progress could not be reset."
		s.settings.failed = true
		slog.Error("Could not reset lesson completion", logErrorKey, msg.err)
		return s, back
	}
	s.lessons.completed = make(map[string]bool)
	s.lessons.progress = progressState{}
	// nothing is left to lose or to read back, so the storage notice no longer applies
	s.lessons.progressNotLoaded = false
	s.lessons.progressNotSaved = false
	s.settings.message = "Lesson progress has been reset."
	s.settings.failed = false
	return s, back
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
