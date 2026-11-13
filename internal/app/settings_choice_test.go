package app

import (
	"errors"
	"fmt"
	"image/color"
	"shellforge/internal/config"
	"shellforge/internal/storage"
	"shellforge/internal/ui"
	"strings"
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
		{"theme with storage", config.ThemeScreen, 2, true, "Theme set to Light.", false, settingsThemeIndex},
		{"theme without storage", config.ThemeScreen, 2, false, "Theme set to Light until Shellforge exits; settings storage is unavailable.", true, settingsThemeIndex},
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
				back = len(ui.Themes) + 1
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

func TestBackgroundColorAppliesThemeAndStoresAutoOnce(t *testing.T) {
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
				t.Fatal("no save command returned for the first detection")
			}
			result = updateModel(t, result, command())
			if len(store.savedThemes) != 1 || store.savedThemes[0] != ui.AutoThemeID {
				t.Errorf("saved themes = %v, want [auto] (the detected theme is not persisted)", store.savedThemes)
			}

			_, command = result.Update(tea.BackgroundColorMsg{Color: test.background})
			if command != nil {
				t.Error("a second detection saved again")
			}
		})
	}
}

func TestAutoStoredInDatabaseIsNotWrittenAgain(t *testing.T) {
	restoreDefaultTheme(t)
	store := &fakeStorage{themeID: ui.AutoThemeID, themeFound: true}
	model := NewWithStorage(store)
	if !initRequestsBackground(t, model) {
		t.Error("Init() did not request the background color in Auto mode")
	}

	updated, command := model.Update(tea.BackgroundColorMsg{Color: color.White})

	if got := updated.(State).preferences.theme().ID; got != "light" {
		t.Errorf("theme = %q, want light", got)
	}
	if command != nil || len(store.savedThemes) != 0 {
		t.Error("detection wrote to storage although auto was already stored")
	}
}

func TestChoosingAutoStoresItAndRequestsBackground(t *testing.T) {
	restoreDefaultTheme(t)
	store := &fakeStorage{themeID: "high-contrast", themeFound: true}
	model := NewWithStorage(store)
	model.nav = navigationState{screen: config.ThemeScreen, selection: autoThemeSelection}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.preferences.themeExplicit {
		t.Error("Auto left the theme marked as explicit")
	}
	if got := result.settings.message; got != "Theme set to Auto." {
		t.Errorf("message = %q, want %q", got, "Theme set to Auto.")
	}
	if got := ui.CurrentTheme().ID; got != "high-contrast" {
		t.Errorf("theme = %q before the terminal answers, want it kept", got)
	}
	batch, ok := command().(tea.BatchMsg)
	if !ok {
		t.Fatalf("command returned %T, want tea.BatchMsg", command())
	}
	requested := false
	for _, step := range batch {
		message := step()
		if fmt.Sprintf("%T", message) == "tea.backgroundColorMsg" {
			requested = true
		}
		result = updateModel(t, result, message)
	}
	if !requested {
		t.Error("choosing Auto did not request the terminal background")
	}
	if len(store.savedThemes) != 1 || store.savedThemes[0] != ui.AutoThemeID {
		t.Errorf("saved themes = %v, want [auto]", store.savedThemes)
	}

	result = updateModel(t, result, tea.BackgroundColorMsg{Color: color.White})
	if got := result.preferences.theme().ID; got != "light" {
		t.Errorf("theme after the reply = %q, want light", got)
	}
}

func TestAutoWithoutStorageStillRequestsBackground(t *testing.T) {
	restoreDefaultTheme(t)
	model := New()
	model.preferences.themeExplicit = true
	model.nav = navigationState{screen: config.ThemeScreen, selection: autoThemeSelection}

	_, command := model.Update(keyPress(tea.KeyEnter, ""))

	if command == nil {
		t.Fatal("choosing Auto without storage returned no command")
	}
}

func TestSettingsShowsAutoWithResolvedTheme(t *testing.T) {
	restoreDefaultTheme(t)
	model := New()
	model = updateModel(t, model, tea.BackgroundColorMsg{Color: color.White})

	if got := model.themeSettingLabel(); got != "Auto (Light)" {
		t.Errorf("label = %q, want Auto (Light)", got)
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
		model.nav = navigationState{screen: config.ThemeScreen, selection: 3} // High contrast
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

func resetSettingsConfirmation(t *testing.T, store storage.Storage, selection int) (State, tea.Cmd) {
	t.Helper()
	model := New()
	if store != nil {
		model = NewWithStorage(store)
	}
	model.nav = navigationState{screen: config.SettingsScreen, selection: settingsResetSettingsIndex}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != config.ResetSettingsScreen || model.nav.selection != 0 {
		t.Fatalf("after choosing Reset settings: screen %d selection %d, want the confirmation on No", model.nav.screen, model.nav.selection)
	}
	model.nav.selection = selection

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	return updated.(State), command
}

func TestResetSettingsDefaultChoiceChangesNothing(t *testing.T) {
	restoreDefaultTheme(t)
	store := &fakeStorage{styleID: "bar-steady", styleFound: true, themeID: "light", themeFound: true}

	result, command := resetSettingsConfirmation(t, store, 0)

	if command != nil || store.resetSettingsCalls != 0 {
		t.Error("choosing No reset the settings")
	}
	if result.nav.screen != config.SettingsScreen || result.nav.selection != settingsResetSettingsIndex {
		t.Errorf("after No: screen %d selection %d, want Settings on Reset settings", result.nav.screen, result.nav.selection)
	}
	if result.preferences.theme().ID != "light" || result.preferences.cursorStyle().ID != "bar-steady" {
		t.Error("choosing No changed a setting")
	}
}

func TestResetSettingsRestoresDefaultsAndRedetectsTheme(t *testing.T) {
	restoreDefaultTheme(t)
	store := &fakeStorage{styleID: "bar-steady", styleFound: true, themeID: "high-contrast", themeFound: true}

	result, command := resetSettingsConfirmation(t, store, 1)

	if got := result.preferences.cursorStyle().ID; got != ui.DefaultCursorStyle.ID {
		t.Errorf("cursor style = %q, want the default", got)
	}
	if result.preferences.themeExplicit || ui.CurrentTheme().ID != ui.DefaultTheme.ID {
		t.Errorf("theme not back to the default Auto start: explicit %v, active %q", result.preferences.themeExplicit, ui.CurrentTheme().ID)
	}
	if result.settings.message != "Settings have been reset." || result.settings.failed {
		t.Errorf("message = %q (failed %v)", result.settings.message, result.settings.failed)
	}
	if result.nav.screen != config.SettingsScreen || result.nav.selection != settingsResetSettingsIndex {
		t.Errorf("after reset: screen %d selection %d, want Settings on Reset settings", result.nav.screen, result.nav.selection)
	}

	batch, ok := command().(tea.BatchMsg)
	if !ok {
		t.Fatalf("command returned %T, want tea.BatchMsg", command())
	}
	requested := false
	for _, step := range batch {
		message := step()
		if fmt.Sprintf("%T", message) == "tea.backgroundColorMsg" {
			requested = true
		}
		result = updateModel(t, result, message)
	}
	if store.resetSettingsCalls != 1 {
		t.Errorf("ResetSettings calls = %d, want 1", store.resetSettingsCalls)
	}
	if !requested {
		t.Error("reset did not ask the terminal for its background")
	}

	result = updateModel(t, result, tea.BackgroundColorMsg{Color: color.White})
	if got := result.preferences.theme().ID; got != "light" {
		t.Errorf("theme after detection = %q, want light", got)
	}
}

func TestResetSettingsWithoutStorageResetsForTheSession(t *testing.T) {
	restoreDefaultTheme(t)

	result, command := resetSettingsConfirmation(t, nil, 1)

	if command == nil {
		t.Error("no command returned; the terminal background should still be requested")
	}
	if !result.settings.failed || !strings.Contains(result.settings.message, "storage is unavailable") {
		t.Errorf("settings without storage = %#v", result.settings)
	}
}

func TestFailedSettingsResetKeepsDefaultsAndReportsFailure(t *testing.T) {
	restoreDefaultTheme(t)
	store := &fakeStorage{settingsResetErr: errors.New("database unavailable")}

	result, command := resetSettingsConfirmation(t, store, 1)
	batch := command().(tea.BatchMsg)
	for _, step := range batch {
		result = updateModel(t, result, step())
	}

	if !result.settings.failed || !strings.Contains(result.settings.message, "could not be reset in storage") {
		t.Errorf("settings after failed reset = %#v", result.settings)
	}
	if result.preferences.themeExplicit {
		t.Error("a failed storage reset undid the in-memory reset")
	}
}
