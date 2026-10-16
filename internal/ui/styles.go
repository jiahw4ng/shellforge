package ui

import "charm.land/lipgloss/v2"

const (
	// AccentColorCode is the shared ANSI color for headings and controls.
	AccentColorCode = "33"
	// TextColorCode is the shared ANSI color for explanatory text.
	TextColorCode = "252"
	// MutedColorCode is the shared ANSI color for secondary information.
	MutedColorCode = "250"

	whiteColorCode  = "15"
	redColorCode    = "196"
	greenColorCode  = "42"
	orangeColorCode = "208"
)

var (
	AccentColor = lipgloss.Color(AccentColorCode)
	TextColor   = lipgloss.Color(TextColorCode)
	MutedColor  = lipgloss.Color(MutedColorCode)
	WhiteColor  = lipgloss.Color(whiteColorCode)
	RedColor    = lipgloss.Color(redColorCode)
	GreenColor  = lipgloss.Color(greenColorCode)
	OrangeColor = lipgloss.Color(orangeColorCode)
)

// styles for the TUI
var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Underline(true)
	ControlStyle  = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	TextStyle     = lipgloss.NewStyle().Foreground(TextColor)
	ErrorStyle    = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	SuccessStyle  = lipgloss.NewStyle().Foreground(GreenColor)
	FailureStyle  = lipgloss.NewStyle().Foreground(RedColor)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	MutedStyle    = lipgloss.NewStyle().Foreground(MutedColor)
	HintStyle     = TextStyle
	WarningStyle  = lipgloss.NewStyle().Foreground(OrangeColor)
	HelpStyle     = TextStyle
)
