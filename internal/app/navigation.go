package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation updates non-terminal navigation and reports whether the app should quit.
func (s *State) handleNavigation(msg tea.KeyMsg) bool {
	switch {
	case key.Matches(msg, config.Keys.Quit):
		return true
	case key.Matches(msg, config.Keys.Up):
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		if s.exitSelected() {
			return true
		}
		s.handleEnter()
	}

	return false
}

func (s State) exitSelected() bool {
	return s.nav.screen == config.MenuScreen && s.nav.selection == len(config.MenuItems)-1
}

func (s State) selectableItemCount() int {
	switch s.nav.screen {
	case config.MenuScreen:
		return len(config.MenuItems)
	case config.LessonsScreen:
		if len(s.lessons.available) == 0 && s.lessons.err == nil {
			return 0
		}
		return len(s.lessons.available) + 1
	case config.SettingsScreen:
		return len(config.SettingsItems)
	case config.CursorStyleScreen:
		return len(ui.CursorStyles) + 1
	case config.ThemeScreen:
		return len(ui.Themes) + 2 // Auto, every theme, then Back
	case config.ResetConfirmationScreen:
		return len(config.ResetConfirmationItems)
	case config.ResetSettingsScreen:
		return len(config.ResetSettingsItems)
	case config.ManageSandboxesScreen:
		return s.manageSandboxesScreen().ItemCount()
	case config.DeleteSandboxesScreen:
		return 2 // No, go back; Yes, delete
	default:
		return 0
	}
}

func (s *State) moveSelection(delta int) {
	next := s.nav.selection + delta
	if next >= 0 && next < s.selectableItemCount() {
		s.nav.selection = next
	}
}

// handleEnter updates the application state for the selected item on the active screen.
func (s *State) handleEnter() {
	switch s.nav.screen {
	case config.MenuScreen:
		s.activateMenuSelection()
	case config.LessonsScreen:
		s.activateLessonSelection()
	case config.SettingsScreen:
		s.activateSettingsSelection()
	case config.ResetConfirmationScreen:
		s.activateResetConfirmationSelection()
	case config.ResetSettingsScreen:
		s.returnToSettings(settingsResetSettingsIndex)
	case config.ManageSandboxesScreen:
		// Leaving stops the listing: its spinner halts and a late result is ignored.
		s.settings.sandboxes.isLoading = false
		s.returnToSettings(settingsManageSandboxesIndex)
	default:
		s.nav.screen = config.MenuScreen
	}
}

func (s *State) activateMenuSelection() {
	switch s.nav.selection {
	case 0:
		s.nav.screen = config.SandboxTerminalScreen
	case 1:
		s.nav.screen = config.LessonsScreen
		s.nav.selection = 0
	case 2:
		s.nav.screen = config.SettingsScreen
		s.nav.selection = 0
		s.settings = settingsState{}
	}
}

func (s *State) activateLessonSelection() {
	if len(s.lessons.available) == 0 {
		if s.lessons.err != nil {
			s.returnToMainMenu()
		}
		return
	}
	if s.nav.selection == len(s.lessons.available) {
		s.returnToMainMenu()
		return
	}

	s.openLesson(s.nav.selection)
}

// openLesson shows a lesson from its first page, with no hints revealed.
func (s *State) openLesson(index int) {
	s.lessons.activeIdx = index
	s.lessons.activePage = 0
	s.lessons.revealedHints = 0
	s.lessons.hintsExpanded = false
	s.lessons.guide.GotoTop()
	s.nav.screen = config.LessonTerminalScreen
}

const (
	settingsCursorStyleIndex = iota
	settingsThemeIndex
	settingsManageSandboxesIndex
	settingsResetIndex
	settingsResetSettingsIndex
)

func (s *State) activateSettingsSelection() {
	switch s.nav.selection {
	case settingsCursorStyleIndex:
		s.nav.screen = config.CursorStyleScreen
		s.nav.selection = s.currentCursorStyleIndex()
	case settingsThemeIndex:
		s.nav.screen = config.ThemeScreen
		s.nav.selection = s.currentThemeIndex()
	case settingsManageSandboxesIndex:
		s.nav.screen = config.ManageSandboxesScreen
		s.nav.selection = 0
	case settingsResetIndex:
		s.nav.screen = config.ResetConfirmationScreen
		s.nav.selection = 0
	case settingsResetSettingsIndex:
		s.nav.screen = config.ResetSettingsScreen
		s.nav.selection = 0
	default:
		s.returnToMainMenu()
	}
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

// returnToSettings goes back to the Settings screen with the given item selected.
func (s *State) returnToSettings(item int) {
	s.nav.screen = config.SettingsScreen
	s.nav.selection = item
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

// resetSettings restores the default cursor style and Auto theme, returns to
// Settings, and clears the saved settings when storage is available. Auto
// starts from the default theme and asks the terminal for its background
// again, exactly as on a first launch. Lesson completions are left alone.
func (s *State) resetSettings() tea.Cmd {
	s.preferences.cursorStyleID = ""
	s.preferences.themeExplicit = false
	s.preferences.autoStored = false
	s.setTheme(ui.DefaultTheme)
	s.returnToSettings(settingsResetSettingsIndex)

	if s.preferences.store == nil {
		s.settings.message = "Settings were reset until Shellforge exits; settings storage is unavailable."
		s.settings.failed = true
		return tea.RequestBackgroundColor
	}
	s.settings.message = "Settings have been reset."
	s.settings.failed = false
	return tea.Batch(resetStoredSettings(s.preferences.store), tea.RequestBackgroundColor)
}

// activateAutoTheme switches to Auto mode. The current theme stays until the
// terminal answers, so a terminal that never replies keeps what is on screen.
func (s *State) activateAutoTheme() tea.Cmd {
	s.preferences.themeExplicit = false
	s.preferences.autoStored = s.preferences.store != nil
	save := s.reportSettingChosen(config.SettingsItemTheme, ui.AutoThemeLabel, saveTheme(s.preferences.store, ui.AutoThemeID))
	return tea.Batch(save, tea.RequestBackgroundColor)
}

func (s *State) activateResetConfirmationSelection() {
	if s.nav.selection == 0 {
		s.nav.screen = config.SettingsScreen
		s.nav.selection = 0
	}
}

func (s *State) returnToMainMenu() {
	s.nav.screen = config.MenuScreen
	s.nav.selection = 0
}
