package ui

import "charm.land/lipgloss/v2"

const (
	grayColorCode        = "241"
	whiteColorCode       = "15"
	blueColorCode        = "33"
	redColorCode         = "196"
	greenColorCode       = "42"
	orangeColorCode      = "208"
	lightPurpleColorCode = "183"
)

// HintColorCode is the shared ANSI color for hint text and rendered hint Markdown.
const HintColorCode = "226"

var (
	GrayColor        = lipgloss.Color(grayColorCode)
	WhiteColor       = lipgloss.Color(whiteColorCode)
	BlueColor        = lipgloss.Color(blueColorCode)
	RedColor         = lipgloss.Color(redColorCode)
	GreenColor       = lipgloss.Color(greenColorCode)
	YellowColor      = lipgloss.Color(HintColorCode)
	OrangeColor      = lipgloss.Color(orangeColorCode)
	LightPurpleColor = lipgloss.Color(lightPurpleColorCode)
)

// styles for the TUI
var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Underline(true)
	ErrorStyle    = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	SuccessStyle  = lipgloss.NewStyle().Foreground(GreenColor)
	FailureStyle  = lipgloss.NewStyle().Foreground(RedColor)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	MutedStyle    = lipgloss.NewStyle().Foreground(GrayColor)
	HintStyle     = lipgloss.NewStyle().Foreground(YellowColor)
	WarningStyle  = lipgloss.NewStyle().Foreground(OrangeColor)
	HelpStyle     = lipgloss.NewStyle().Foreground(LightPurpleColor)
)
