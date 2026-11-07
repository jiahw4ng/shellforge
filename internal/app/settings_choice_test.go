package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The wording below is user-visible, so it is pinned exactly.
func TestChoosingASettingReportsExactWording(t *testing.T) {
	restoreDefaultTheme(t)
	tests := []struct {
		name       string
		screen     config.Screen
		selection  int
		withStore  bool
		want       string
		wantFailed bool
		wantReturn int
	}{
		{"cursor with storage", config.CursorStyleScreen, 3, true, "Cursor style set to Underline (steady).", false, settingsCursorStyleIndex},
		{"cursor without storage", config.CursorStyleScreen, 3, false, "Cursor style set to Underline (steady) until Shellforge exits; settings storage is unavailable.", true, settingsCursorStyleIndex},
		{"theme with storage", config.ThemeScreen, 1, true, "Theme set to Light.", false, settingsThemeIndex},
		{"theme without storage", config.ThemeScreen, 1, false, "Theme set to Light until Shellforge exits; settings storage is unavailable.", true, settingsThemeIndex},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := New()
			if test.withStore {
				model = NewWithStorage(&fakeStorage{})
			}
			model.nav = navigationState{screen: test.screen, selection: test.selection}

			updated, command := model.Update(keyPress(tea.KeyEnter, ""))
			result := updated.(State)

			if result.settings.message != test.want || result.settings.failed != test.wantFailed {
				t.Errorf("message = %q (failed %v), want %q (failed %v)", result.settings.message, result.settings.failed, test.want, test.wantFailed)
			}
			if result.nav.screen != config.SettingsScreen || result.nav.selection != test.wantReturn {
				t.Errorf("returned to screen %d item %d, want Settings item %d", result.nav.screen, result.nav.selection, test.wantReturn)
			}
			if saves := command != nil; saves != test.withStore {
				t.Errorf("schedules a save = %v, want %v", saves, test.withStore)
			}
		})
	}
}

func TestChoosingBackLeavesSettingsUntouched(t *testing.T) {
	restoreDefaultTheme(t)
	for name, screen := range map[string]config.Screen{"cursor": config.CursorStyleScreen, "theme": config.ThemeScreen} {
		t.Run(name, func(t *testing.T) {
			back := len(ui.CursorStyles)
			if screen == config.ThemeScreen {
				back = len(ui.Themes)
			}
			model := NewWithStorage(&fakeStorage{})
			model.nav = navigationState{screen: screen, selection: back}

			updated, command := model.Update(keyPress(tea.KeyEnter, ""))
			result := updated.(State)

			if command != nil || result.settings.message != "" || result.nav.screen != config.SettingsScreen {
				t.Fatalf("Back: command %v, message %q, screen %d; want no work, no message, Settings", command != nil, result.settings.message, result.nav.screen)
			}
		})
	}
}
