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
	s.showInSettings(childReport{message: "Cursor style could not be saved; it applies until Shellforge exits.", failed: true})
	return s, nil
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

// cursorStyleBehaviour is the Cursor style screen's behaviour: a list of every
// cursor style, then Back.
func cursorStyleBehaviour() screenBehaviour {
	return screenBehaviour{
		parent: new(config.SettingsScreen),
		// The screen opens on the cursor style in use.
		onArrival: func(s *State) tea.Cmd {
			s.nav.selection = s.currentCursorStyleIndex()
			return nil
		},
		list: &listBehaviour{
			// Every style, then Back.
			itemCount: func(State) int { return len(ui.CursorStyles) + 1 },
			// Enter applies the selected cursor style, reports it to Settings,
			// and schedules persistence when storage is available. Back only
			// returns.
			onSelectItem: func(s *State) tea.Cmd {
				selected := s.nav.selection
				if selected >= len(ui.CursorStyles) {
					return s.goBackToParentScreen()
				}

				style := ui.CursorStyles[selected]
				s.preferences.cursorStyleID = style.ID
				report, save := s.settingChosen(config.SettingsItemCursorStyle, style.Label, saveCursorStyle(s.preferences.store, style.ID))
				return tea.Batch(s.goBackWithReport(report), save)
			},
		},
		// The view lists the cursor styles and marks the current one.
		view: func(s State, helpWidth int) string {
			return screens.CursorStyle(cursorStyleLabels(), s.currentCursorStyleIndex(), s.nav.selection, s.navigationHelp(helpWidth))
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}
