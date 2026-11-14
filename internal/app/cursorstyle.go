package app

import (
	"context"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/storage"
	"shellforge/internal/ui"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type cursorStyleSavedMsg struct {
	styleID string
	err     error
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

// handleCursorStyleKeyPress delegates ordinary navigation and applies the
// selected cursor style or Back on Enter.
func (s State) handleCursorStyleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if !key.Matches(msg, config.Keys.SelectItem) {
		return s.handleNavigationKeyPress(msg)
	}
	return s, s.activateCursorStyleSelection()
}

func (s State) currentCursorStyleIndex() int {
	current := s.preferences.cursorStyle()
	for index, style := range ui.CursorStyles {
		if style.ID == current.ID {
			return index
		}
	}
	return 0
}

// activateCursorStyleSelection applies the selected cursor style, returns to
// Settings, and schedules persistence when storage is available.
func (s *State) activateCursorStyleSelection() tea.Cmd {
	selected := s.nav.selection
	s.returnToSettings(settingsCursorStyleIndex)
	if selected >= len(ui.CursorStyles) {
		return nil
	}

	style := ui.CursorStyles[selected]
	s.preferences.cursorStyleID = style.ID
	return s.reportSettingChosen(config.SettingsItemCursorStyle, style.Label, saveCursorStyle(s.preferences.store, style.ID))
}

func saveCursorStyle(store storage.SettingsStore, styleID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.UpdateCursorStyle(ctx, styleID)
		return cursorStyleSavedMsg{styleID: styleID, err: err}
	}
}

func cursorStyleLabels() []string {
	labels := make([]string, len(ui.CursorStyles))
	for index, style := range ui.CursorStyles {
		labels[index] = style.Label
	}
	return labels
}
