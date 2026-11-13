package app

import (
	"fmt"
	"image/color"
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

func initRequestsBackground(t *testing.T, model State) bool {
	t.Helper()
	batch, ok := model.Init()().(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() returned %T, want tea.BatchMsg", model.Init()())
	}
	for _, command := range batch {
		// tea.RequestBackgroundColor's message type is unexported, so match by name.
		if fmt.Sprintf("%T", command()) == "tea.backgroundColorMsg" {
			return true
		}
	}
	return false
}

func TestInitRequestsBackgroundOnlyWhenNoThemeIsChosen(t *testing.T) {
	restoreDefaultTheme(t)
	if !initRequestsBackground(t, NewWithStorage(&fakeStorage{})) {
		t.Error("Init() did not request the background color with no saved theme")
	}
	if !initRequestsBackground(t, New()) {
		t.Error("Init() did not request the background color without storage")
	}
	saved := NewWithStorage(&fakeStorage{themeID: "light", themeFound: true})
	if initRequestsBackground(t, saved) {
		t.Error("Init() requested the background color despite a saved theme")
	}
	unknown := NewWithStorage(&fakeStorage{themeID: "no-such-theme", themeFound: true})
	if !initRequestsBackground(t, unknown) {
		t.Error("Init() did not request the background color for an unknown saved theme")
	}
}

func TestBackgroundColorChoosesAndSavesTheme(t *testing.T) {
	tests := []struct {
		name       string
		background color.Color
		want       string
	}{
		{"light terminal", color.White, "light"},
		{"dark terminal", color.Black, "dark"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restoreDefaultTheme(t)
			store := &fakeStorage{}
			model := NewWithStorage(store)

			updated, command := model.Update(tea.BackgroundColorMsg{Color: test.background})
			result := updated.(State)

			if got := result.preferences.theme().ID; got != test.want {
				t.Errorf("theme = %q, want %q", got, test.want)
			}
			if got := ui.CurrentTheme().ID; got != test.want {
				t.Errorf("active theme = %q, want %q", got, test.want)
			}
			if command == nil {
				t.Fatal("no save command returned")
			}
			updateModel(t, result, command())
			if len(store.savedThemes) != 1 || store.savedThemes[0] != test.want {
				t.Errorf("saved themes = %v, want [%s]", store.savedThemes, test.want)
			}
		})
	}
}

func TestBackgroundColorIsIgnoredOnceThemeIsChosen(t *testing.T) {
	t.Run("loaded from storage", func(t *testing.T) {
		restoreDefaultTheme(t)
		store := &fakeStorage{themeID: "high-contrast", themeFound: true}
		model := NewWithStorage(store)

		updated, command := model.Update(tea.BackgroundColorMsg{Color: color.White})

		if got := updated.(State).preferences.theme().ID; got != "high-contrast" {
			t.Errorf("theme = %q, want high-contrast", got)
		}
		if command != nil || len(store.savedThemes) != 0 {
			t.Error("a saved theme was overwritten by detection")
		}
	})
	t.Run("chosen before the reply", func(t *testing.T) {
		restoreDefaultTheme(t)
		store := &fakeStorage{}
		model := NewWithStorage(store)
		model.nav = navigationState{screen: config.ThemeScreen, selection: 2}
		model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

		updated, command := model.Update(tea.BackgroundColorMsg{Color: color.White})

		if got := updated.(State).preferences.theme().ID; got != "high-contrast" {
			t.Errorf("theme = %q, want the learner's high-contrast", got)
		}
		if command != nil {
			t.Error("detection scheduled a save after the learner chose a theme")
		}
	})
}

func TestBackgroundColorWithoutStorageStillAppliesTheme(t *testing.T) {
	restoreDefaultTheme(t)

	updated, command := New().Update(tea.BackgroundColorMsg{Color: color.White})

	if got := updated.(State).preferences.theme().ID; got != "light" {
		t.Errorf("theme = %q, want light", got)
	}
	if command != nil {
		t.Error("a save command was returned without storage")
	}
}
