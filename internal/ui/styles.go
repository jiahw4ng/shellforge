package ui

import "github.com/charmbracelet/lipgloss"

const (
	GrayColor  = lipgloss.Color("241")
	WhiteColor = lipgloss.Color("15")
	BlueColor  = lipgloss.Color("33")
)

// styles for the TUI
var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Underline(true)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	MutedStyle    = lipgloss.NewStyle().Foreground(GrayColor)
)
