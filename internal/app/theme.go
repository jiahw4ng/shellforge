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

type themeSavedMsg struct {
	themeID string
	err     error
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

func saveTheme(store storage.SettingsStore, themeID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.UpdateSaveTheme(ctx, themeID)
		return themeSavedMsg{themeID: themeID, err: err}
	}
}

// currentThemeIndex returns the Theme screen selection for the active mode:
// Auto is the first entry and each theme follows it.
func (s State) currentThemeIndex() int {
	if !s.preferences.themeExplicit {
		return autoThemeSelection
	}
	current := s.preferences.theme()
	for index, theme := range ui.Themes {
		if theme.ID == current.ID {
			return index + 1
		}
	}
	return autoThemeSelection
}

// setTheme makes theme the active theme. Styles the help model copied when it
// was created are rebuilt, because Lipgloss styles are values.
func (s *State) setTheme(theme ui.Theme) {
	s.preferences.themeID = theme.ID
	ui.ApplyTheme(theme)
	applyHelpStyles(&s.help)
}

// autoThemeSelection is the Theme screen's first entry; ui.Themes follow it.
const autoThemeSelection = 0

// activateThemeSelection applies the selected theme (or Auto mode), returns to
// Settings, and schedules persistence when storage is available. Choosing Auto
// also asks the terminal for its background again.
func (s *State) activateThemeSelection() tea.Cmd {
	selected := s.nav.selection
	s.returnToSettings(settingsThemeIndex)
	if selected == autoThemeSelection {
		return s.activateAutoTheme()
	}
	if selected > len(ui.Themes) {
		return nil
	}

	theme := ui.Themes[selected-1]
	s.setTheme(theme)
	s.preferences.themeExplicit = true
	s.preferences.autoStored = false
	return s.reportSettingChosen(config.SettingsItemTheme, theme.Label, saveTheme(s.preferences.store, theme.ID))
}

// handleThemeKeyPress delegates ordinary navigation and applies the selected
// theme or Back on Enter.
func (s State) handleThemeKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if !key.Matches(msg, config.Keys.SelectItem) {
		return s.handleNavigationKeyPress(msg)
	}
	return s, s.activateThemeSelection()
}

// activateAutoTheme switches to Auto mode. The current theme stays until the
// terminal answers, so a terminal that never replies keeps what is on screen.
func (s *State) activateAutoTheme() tea.Cmd {
	s.preferences.themeExplicit = false
	s.preferences.autoStored = s.preferences.store != nil
	save := s.reportSettingChosen(config.SettingsItemTheme, ui.AutoThemeLabel, saveTheme(s.preferences.store, ui.AutoThemeID))
	return tea.Batch(save, tea.RequestBackgroundColor)
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

// themeLabels lists Auto first, then every theme, matching Theme screen selection.
func themeLabels() []string {
	labels := []string{ui.AutoThemeLabel}
	for _, theme := range ui.Themes {
		labels = append(labels, theme.Label)
	}
	return labels
}

// lightThemeID is the theme detected on a light terminal background.
const lightThemeID = "light"
