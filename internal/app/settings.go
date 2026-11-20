package app

import (
	"fmt"
	"shellforge/internal/config"
	"shellforge/internal/screens"
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

// settingChosen is the report for a setting the learner just chose, and the
// command that saves it. Without storage the choice only lasts until
// Shellforge exits, so it is reported as a failure and nothing is saved.
func (s State) settingChosen(name, label string, save tea.Cmd) (*childReport, tea.Cmd) {
	if s.preferences.store == nil {
		return &childReport{message: name + " set to " + label + " until Shellforge exits; settings storage is unavailable.", failed: true}, nil
	}
	return &childReport{message: name + " set to " + label + "."}, save
}

// showInSettings sets the Settings status line, which is only written here.
// Settings' onReturn shows its children's reports with it, and the handlers
// for saves that finish after their screen closed call it directly.
func (s *State) showInSettings(report childReport) {
	s.settings.message = report.message
	s.settings.failed = report.failed
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

// settingsBehaviour is the Settings screen's behaviour. Its view lists the
// settings with their current values.
func settingsBehaviour() screenBehaviour {
	return screenBehaviour{
		parent: new(config.MainMenuScreen),
		list: &listBehaviour{
			itemCount: func(State) int { return len(config.SettingsItems) },
			// Enter opens the chosen setting's screen, or returns to the main
			// menu from Back. A screen with a current value selects it itself.
			onSelectItem: func(s *State) tea.Cmd {
				switch s.nav.selection {
				case settingsCursorStyleIndex:
					return s.openScreen(navigationState{screen: config.ChooseCursorStyleScreen})
				case settingsThemeIndex:
					return s.openScreen(navigationState{screen: config.ChooseThemeScreen})
				case settingsManageSandboxesIndex:
					return s.openScreen(navigationState{screen: config.ManageSandboxesScreen})
				case settingsResetIndex:
					return s.openScreen(navigationState{screen: config.ResetLessonProgressScreen})
				case settingsResetSettingsIndex:
					return s.openScreen(navigationState{screen: config.ResetSettingsScreen})
				default:
					return s.goBack()
				}
			},
		},
		// Opening Settings from the main menu clears its last message.
		// Returning from a child does not run this, so a report stays.
		onArrival: func(s *State, _ config.Screen) tea.Cmd {
			s.settings = settingsState{}
			return nil
		},
		// A child that changed or reset something reports it here, and the
		// status line shows it. A plain Back keeps the last message.
		onReturn: func(s *State, report *childReport) tea.Cmd {
			if report != nil {
				s.showInSettings(*report)
			}
			return nil
		},
		view: func(s State, helpWidth int) string {
			return screens.Settings(s.settingsItems(), s.nav.selection, s.settings.message, s.settings.failed, s.navigationHelp(helpWidth))
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}
