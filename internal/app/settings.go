package app

import (
	"fmt"
	"shellforge/internal/config"
	"shellforge/internal/ui"
	"slices"

	tea "charm.land/bubbletea/v2"
)

const (
	settingsCursorStyleIndex = iota
	settingsThemeIndex
	settingsManageSandboxesIndex
	settingsResetIndex
	settingsResetSettingsIndex
)

func (s *State) activateSettingsSelection() tea.Cmd {
	switch s.nav.selection {
	case settingsCursorStyleIndex:
		return s.navigate(config.CursorStyleScreen, s.currentCursorStyleIndex())
	case settingsThemeIndex:
		return s.navigate(config.ThemeScreen, s.currentThemeIndex())
	case settingsManageSandboxesIndex:
		return s.navigate(config.ManageSandboxesScreen, 0)
	case settingsResetIndex:
		return s.navigate(config.ResetConfirmationScreen, 0)
	case settingsResetSettingsIndex:
		return s.navigate(config.ResetSettingsScreen, 0)
	default:
		return s.returnToMainMenu()
	}
}

// returnToSettings goes back to the Settings screen with the given item selected.
func (s *State) returnToSettings(item int) tea.Cmd {
	return s.navigate(config.SettingsScreen, item)
}

// reportSettingChosen tells the learner what happened to a setting they just
// chose. Without storage the choice only lasts until Shellforge exits, so it is
// reported as a failure and nothing is saved; otherwise save is returned to run.
func (s *State) reportSettingChosen(name, label string, save tea.Cmd) tea.Cmd {
	if s.preferences.store == nil {
		s.settings.message = name + " set to " + label + " until Shellforge exits; settings storage is unavailable."
		s.settings.failed = true
		return nil
	}
	s.settings.message = name + " set to " + label + "."
	s.settings.failed = false
	return save
}

// settingsItems labels the preference items with the current choices, padding
// each "Name:" to the widest one so the values start in the same column.
func (s State) settingsItems() []string {
	items := slices.Clone(config.SettingsItems)
	values := map[int]string{
		settingsCursorStyleIndex: s.preferences.cursorStyle().Label,
		settingsThemeIndex:       s.themeSettingLabel(),
	}
	nameWidth := 0
	for index := range values {
		nameWidth = max(nameWidth, len(items[index])+len(settingNameSuffix))
	}
	for index, value := range values {
		items[index] = fmt.Sprintf("%-*s %s", nameWidth, items[index]+settingNameSuffix, value)
	}
	return items
}

// themeSettingLabel names the theme mode in Settings. Auto also shows the
// theme it currently resolves to.
func (s State) themeSettingLabel() string {
	if s.preferences.themeExplicit {
		return s.preferences.theme().Label
	}
	return ui.AutoThemeLabel + " (" + s.preferences.theme().Label + ")"
}

// settingNameSuffix separates a setting's name from its current value.
const settingNameSuffix = ":"
