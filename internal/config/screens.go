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
)

const (
	menuItemSandbox         = "Sandbox"
	menuItemChooseLesson    = "Choose lesson"
	menuItemExit            = "Exit"
	settingsItemReset       = "Reset lesson progress"
	SettingsItemCursorStyle = "Cursor style"
	resetItemCancel         = "No, go back"
	resetItemConfirm        = "Yes, reset lesson progress"
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
	settingsItemReset,
	BackMenuLabel,
}

var ResetConfirmationItems = []string{
	resetItemCancel,
	resetItemConfirm,
}
