package ui

import "charm.land/lipgloss/v2"

var (
	// grayColor colors muted interface elements.
	grayColor = lipgloss.Color("241")
	// WhiteColor colors selected interface elements and borders.
	WhiteColor = lipgloss.Color("15")
	// BlueColor colors titles and active pagination markers.
	BlueColor = lipgloss.Color("33")
	// redColor colors errors and failed checks.
	redColor = lipgloss.Color("196")
	// greenColor colors successful checks.
	greenColor = lipgloss.Color("42")
	// yellowColor colors lesson hints.
	yellowColor = lipgloss.Color("226")
	// lightPurpleColor colors help text.
	lightPurpleColor = lipgloss.Color("183")
)

var (
	// TitleStyle emphasizes screen and lesson titles.
	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Underline(true)
	// ErrorStyle emphasizes error headings.
	ErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(redColor).Underline(true)
	// SuccessStyle colors successful outcomes.
	SuccessStyle = lipgloss.NewStyle().Foreground(greenColor)
	// FailureStyle colors failed outcomes.
	FailureStyle = lipgloss.NewStyle().Foreground(redColor)
	// SelectedStyle emphasizes the active menu item.
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	// MutedStyle colors secondary interface text.
	MutedStyle = lipgloss.NewStyle().Foreground(grayColor)
	// HintStyle colors lesson hints.
	HintStyle = lipgloss.NewStyle().Foreground(yellowColor)
	// HelpStyle colors shortcut help.
	HelpStyle = lipgloss.NewStyle().Foreground(lightPurpleColor)
)
