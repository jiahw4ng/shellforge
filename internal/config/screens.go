// Package config
package config

type Screen int

const (
	MainMenuScreen Screen = iota
	LessonsScreen
	LessonTerminalScreen
	SettingsScreen
	ResetLessonProgressScreen
	SandboxTerminalScreen
	ChooseCursorStyleScreen
	ChooseThemeScreen
	ResetSettingsScreen
	ManageSandboxesScreen
	DeleteSandboxesScreen
	// add more on top of this
	TotalScreenCount
)

const (
	menuItemSandbox         = "Sandbox"
	menuItemChooseLesson    = "Choose lesson"
	menuItemExit            = "Exit"
	settingsItemReset       = "Reset lesson progress"
	settingsItemResetAll    = "Reset settings"
	SettingsItemCursorStyle = "Cursor style"
	SettingsItemTheme       = "Theme"
	// SettingsItemManageSandboxes opens the list of every Shellforge sandbox.
	SettingsItemManageSandboxes = "Manage sandboxes"
	resetItemCancel             = "No, go back"
	resetItemConfirm            = "Yes, reset"
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
	SettingsItemManageSandboxes,
	settingsItemReset,
	settingsItemResetAll,
	BackMenuLabel,
}

var ResetConfirmationItems = []string{
	resetItemCancel,
	resetItemConfirm,
}
