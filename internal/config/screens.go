// Package config todo
package config

type Screen int

const (
	MenuScreen Screen = iota
	LessonsScreen
	LessonTerminalScreen
	SettingsScreen
	ResetConfirmationScreen
	SandboxTerminalScreen
	CursorStyleScreen
	ThemeScreen
	ResetSettingsScreen
)

const (
	menuItemSandbox         = "Sandbox"
	menuItemChooseLesson    = "Choose lesson"
	menuItemExit            = "Exit"
	settingsItemReset       = "Reset lesson progress"
	settingsItemResetAll    = "Reset settings"
	SettingsItemCursorStyle = "Cursor style"
	SettingsItemTheme       = "Theme"
	resetItemCancel         = "No, go back"
	resetItemConfirm        = "Yes, reset lesson progress"
	resetSettingsConfirm    = "Yes, reset settings"
)

// SettingsMenuLabel is the shared label for the Settings destination and title.
const SettingsMenuLabel = "Settings"

// BackMenuLabel is the shared label for navigation back from a submenu.
const BackMenuLabel = "Back"

var MenuItems = []string{
	menuItemSandbox,
	menuItemChooseLesson,
	SettingsMenuLabel,
	menuItemExit,
}

var SettingsItems = []string{
	SettingsItemCursorStyle,
	SettingsItemTheme,
	settingsItemReset,
	settingsItemResetAll,
	BackMenuLabel,
}

var ResetConfirmationItems = []string{
	resetItemCancel,
	resetItemConfirm,
}

// ResetSettingsItems are the choices on the settings reset confirmation, with
// the safe one first.
var ResetSettingsItems = []string{
	resetItemCancel,
	resetSettingsConfirm,
}
