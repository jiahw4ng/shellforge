package tui

import "github.com/charmbracelet/lipgloss"

const (
	grayColor  = lipgloss.Color("241")
	whiteColor = lipgloss.Color("15")
	blueColor  = lipgloss.Color("33")
)

// styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(blueColor).Underline(true)
	selected   = lipgloss.NewStyle().Bold(true).Foreground(whiteColor)
	muted      = lipgloss.NewStyle().Foreground(grayColor)
)
