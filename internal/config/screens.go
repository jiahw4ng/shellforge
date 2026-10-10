// Package config defines Shellforge's screen, menu, and key-binding vocabulary.
package config

// Screen identifies one application navigation state.
type Screen int

// MenuScreen and the other Screen constants identify Shellforge's navigable
// application states.
const (
	MenuScreen Screen = iota
	LessonsScreen
	LessonTerminalScreen
	SettingsScreen
	ResetConfirmationScreen
	SandboxTerminalScreen
)

// MenuItems lists the choices shown on the main menu.
var MenuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
	"Exit",
}

// SettingsItems lists the choices shown on the settings screen.
var SettingsItems = []string{
	"Reset lesson progress",
	"Back",
}

// ResetConfirmationItems lists the choices shown before completion data is reset.
var ResetConfirmationItems = []string{
	"No, go back",
	"Yes, reset lesson progress",
}
