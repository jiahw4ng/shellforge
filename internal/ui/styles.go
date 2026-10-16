package ui

import "charm.land/lipgloss/v2"

const (
	// AccentColorCode is the shared ANSI color for headings and controls.
	AccentColorCode = "33"
	// TextColorCode is the shared ANSI color for explanatory text.
	TextColorCode = "252"
	// MutedColorCode is the shared ANSI color for secondary information.
	MutedColorCode = "250"

	WhiteColorCode  = "15"
	RedColorCode    = "196"
	GreenColorCode  = "42"
	OrangeColorCode = "208"
)

var (
	AccentColor = lipgloss.Color(AccentColorCode)
	TextColor   = lipgloss.Color(TextColorCode)
	MutedColor  = lipgloss.Color(MutedColorCode)
	WhiteColor  = lipgloss.Color(WhiteColorCode)
	RedColor    = lipgloss.Color(RedColorCode)
	GreenColor  = lipgloss.Color(GreenColorCode)
	OrangeColor = lipgloss.Color(OrangeColorCode)
)

// styles for the TUI
var (
	TitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Underline(true)
	ControlStyle     = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	PrimaryTextStyle = lipgloss.NewStyle().Foreground(WhiteColor)
	TextStyle        = lipgloss.NewStyle().Foreground(TextColor)
	ErrorStyle       = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	SuccessStyle     = lipgloss.NewStyle().Foreground(GreenColor)
	FailureStyle     = lipgloss.NewStyle().Foreground(RedColor)
	SelectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	MutedStyle       = lipgloss.NewStyle().Foreground(MutedColor)
	HintStyle        = TextStyle
	WarningStyle     = lipgloss.NewStyle().Foreground(OrangeColor)
	HelpStyle        = TextStyle
)
