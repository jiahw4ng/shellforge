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

// lessonItems progress from basic shell navigation to more advanced Unix work.
var lessonItems = []string{
	"Getting around",
	"Files and directories",
	"Viewing and searching files",
	"Copying, moving, and removing files",
	"Permissions and ownership",
	"Processes and job control",
	"Pipes and redirection",
	"Shell variables and scripts",
	"Text processing",
	"System investigation",
}
