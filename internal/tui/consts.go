package tui

import (
	"shellforge/internal/container"

	"github.com/charmbracelet/lipgloss"
	bubbleterm "github.com/taigrr/bubbleterm"
)

type currentScreen int

// screens that the user can navigate to in the TUI
const (
	menuScreen currentScreen = iota
	lessonsScreen
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

const (
	grayColor  = lipgloss.Color("241")
	whiteColor = lipgloss.Color("15")
)

// styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(whiteColor).Underline(true)
	selected   = lipgloss.NewStyle().Bold(true).Foreground(whiteColor)
	muted      = lipgloss.NewStyle().Foreground(grayColor)
)

// Model is Shellforge's initial navigation state
type Model struct {
	currentScreen    currentScreen
	selectedOption   int
	selectedLesson   int
	featureReturnTo  currentScreen
	termWidth        int
	termHeight       int
	terminal         *bubbleterm.Model
	lessonContainer  *container.LessonContainer
	terminalExit     <-chan struct{}
	terminalStartErr error
}
