package tui

type currentScreen int

// screens that the user can navigate to in the TUI
const (
	menuScreen currentScreen = iota
	lessonsScreen
	lessonScreen
	featureScreen
	terminalScreen
)

// options that the user can select from in the main menu
var menuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
}
