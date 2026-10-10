package ui

import "charm.land/lipgloss/v2"

var (
	// GrayColor colors muted interface elements.
	GrayColor = lipgloss.Color("241")
	// WhiteColor colors selected interface elements and borders.
	WhiteColor = lipgloss.Color("15")
	// BlueColor colors titles and active pagination markers.
	BlueColor = lipgloss.Color("33")
	// RedColor colors errors and failed checks.
	RedColor = lipgloss.Color("196")
	// GreenColor colors successful checks.
	GreenColor = lipgloss.Color("42")
	// YellowColor colors lesson hints.
	YellowColor = lipgloss.Color("226")
	// LightPurpleColor colors help text.
	LightPurpleColor = lipgloss.Color("183")
)

var (
	// TitleStyle emphasizes screen and lesson titles.
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Underline(true)
	// ErrorStyle emphasizes error headings.
	ErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	// SuccessStyle colors successful outcomes.
	SuccessStyle = lipgloss.NewStyle().Foreground(GreenColor)
	// FailureStyle colors failed outcomes.
	FailureStyle = lipgloss.NewStyle().Foreground(RedColor)
	// SelectedStyle emphasizes the active menu item.
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	// MutedStyle colors secondary interface text.
	MutedStyle = lipgloss.NewStyle().Foreground(GrayColor)
	// HintStyle colors lesson hints.
	HintStyle = lipgloss.NewStyle().Foreground(YellowColor)
	// HelpStyle colors shortcut help.
	HelpStyle = lipgloss.NewStyle().Foreground(LightPurpleColor)
)
