package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Shared colors for the active theme. ApplyTheme reassigns them.
var (
	AccentColor  color.Color
	TextColor    color.Color
	MutedColor   color.Color
	PrimaryColor color.Color
	RedColor     color.Color
	GreenColor   color.Color
	OrangeColor  color.Color
)

// styles for the TUI, rebuilt from the active theme by ApplyTheme
var (
	TitleStyle       lipgloss.Style
	ControlStyle     lipgloss.Style
	PrimaryTextStyle lipgloss.Style
	TextStyle        lipgloss.Style
	ErrorStyle       lipgloss.Style
	SuccessStyle     lipgloss.Style
	FailureStyle     lipgloss.Style
	SelectedStyle    lipgloss.Style
	MutedStyle       lipgloss.Style
	HintStyle        lipgloss.Style
	WarningStyle     lipgloss.Style
	HelpStyle        lipgloss.Style
)

func applyPalette(palette Palette) {
	AccentColor = lipgloss.Color(palette.Accent)
	TextColor = lipgloss.Color(palette.Text)
	MutedColor = lipgloss.Color(palette.Muted)
	PrimaryColor = lipgloss.Color(palette.Primary)
	RedColor = lipgloss.Color(palette.Red)
	GreenColor = lipgloss.Color(palette.Green)
	OrangeColor = lipgloss.Color(palette.Orange)

	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Underline(true)
	ControlStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	PrimaryTextStyle = lipgloss.NewStyle().Foreground(PrimaryColor)
	TextStyle = lipgloss.NewStyle().Foreground(TextColor)
	ErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(RedColor).Underline(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(GreenColor)
	FailureStyle = lipgloss.NewStyle().Foreground(RedColor)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(AccentColor)
	MutedStyle = lipgloss.NewStyle().Foreground(MutedColor)
	HintStyle = TextStyle
	WarningStyle = lipgloss.NewStyle().Foreground(OrangeColor)
	HelpStyle = TextStyle

	ApplicationFrameStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, ApplicationHorizontalMargin, 0, ApplicationHorizontalMargin)
}
