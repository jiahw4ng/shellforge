package app

import (
	"context"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/screens"
	"shellforge/internal/storage"
	"shellforge/internal/ui"
	"time"

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
	back := s.returnToSettings(settingsResetIndex)
	s.settings.message = "Lesson progress could not be reset because storage is unavailable."
	s.settings.failed = true
	return back
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

// handleCompletionsReset returns to Settings and records either the reset
// failure or a cleared completion and progress state.
func (s State) handleCompletionsReset(msg completionsResetMsg) (tea.Model, tea.Cmd) {
	s.settings.isResetting = false
	back := s.returnToSettings(settingsResetIndex)
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

// resetSettingsBehaviour is the behaviour of the confirmation that restores the
// default settings. Like every confirmation, it lists the safe choice first.
func resetSettingsBehaviour() screenBehaviour {
	return screenBehaviour{
		list: &listBehaviour{
			// Cancel, then confirm.
			itemCount: func(State) int { return len(config.ResetSettingsItems) },
			// Enter restores the default settings on the confirming choice and
			// returns to Settings on the safe one.
			onSelectItem: func(s *State) tea.Cmd {
				if s.nav.selection == confirmItem {
					return s.resetSettings()
				}
				return s.returnToSettings(settingsResetSettingsIndex)
			},
		},
		// The view asks the learner to confirm restoring the default settings.
		view: func(s State, helpWidth int) string {
			return screens.ResetSettingsConfirmation(config.ResetSettingsItems, s.nav.selection, s.navigationHelp(helpWidth))
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}

// resetLessonProgressBehaviour is the behaviour of the confirmation that
// deletes every lesson completion. Like every confirmation, it lists the safe
// choice first.
func resetLessonProgressBehaviour() screenBehaviour {
	return screenBehaviour{
		list: &listBehaviour{
			// Cancel, then confirm.
			itemCount: func(State) int { return len(config.ResetConfirmationItems) },
			// Enter returns to Settings on the safe choice. On the confirming
			// one it reports unavailable storage, or marks the reset as running
			// (which blocks every key but Ctrl+C) and schedules the deletion.
			onSelectItem: func(s *State) tea.Cmd {
				if s.nav.selection != confirmItem {
					return s.returnToSettings(settingsResetIndex)
				}
				if s.lessons.store == nil {
					return s.reportResetUnavailable()
				}
				s.settings.isResetting = true
				return resetLessonCompletions(s.lessons.store)
			},
		},
		// The view hides the help line while the reset runs, because no key
		// but Ctrl+C works then.
		view: func(s State, helpWidth int) string {
			help := ""
			if !s.settings.isResetting {
				help = s.navigationHelp(helpWidth)
			}
			return screens.ResetConfirmation(config.ResetConfirmationItems, s.nav.selection, s.settings.isResetting, help)
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}
