package app

type screen int

const (
	menuScreen screen = iota
	lessonsScreen
	lessonScreen
	settingsScreen
	resetConfirmationScreen
	terminalScreen
)

var menuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
	"Exit",
}

var settingsItems = []string{
	"Reset lesson progress",
	"Back",
}

var resetConfirmationItems = []string{
	"No, go back",
	"Yes, reset lesson progress",
}
