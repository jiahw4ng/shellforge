package ui

import "github.com/charmbracelet/lipgloss"

const (
	GrayColor  = lipgloss.Color("241")
	WhiteColor = lipgloss.Color("15")
	BlueColor  = lipgloss.Color("33")
	RedColor   = lipgloss.Color("196")
	GreenColor = lipgloss.Color("42")
)

// styles for the TUI
var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Underline(true)
	ErrorStyle    = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	SuccessStyle  = lipgloss.NewStyle().Foreground(GreenColor)
	FailureStyle  = lipgloss.NewStyle().Foreground(RedColor)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	MutedStyle    = lipgloss.NewStyle().Foreground(GrayColor)
)
