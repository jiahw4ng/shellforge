package tui

import (
	shellpty "shellforge/internal/pty"

	"github.com/charmbracelet/lipgloss"
)

type screen int

// screens that the user can navigate to in the TUI
const (
	menuScreen screen = iota
	featureScreen
	ptyScreen
)

// options that the user can select from in the main menu
var menuItems = []string{
	"Start learning",
	"Choose lesson",
	"Settings",
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
	screen    screen
	selected  int
	width     int
	height    int
	pty       *shellpty.Session
	ptyOutput string
	ptyError  error
}
