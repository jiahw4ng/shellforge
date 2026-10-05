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
)

var MenuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
	"Exit",
}

var SettingsItems = []string{
	"Reset lesson progress",
	"Back",
}

var ResetConfirmationItems = []string{
	"No, go back",
	"Yes, reset lesson progress",
}
