// Package ui provides shared Lipgloss styles and the application frame.
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	ApplicationBorderSize       = 2
	ApplicationHorizontalMargin = 1
)

var ApplicationFrameStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(WhiteColor).
	Padding(0, ApplicationHorizontalMargin, 0, ApplicationHorizontalMargin)

// ApplicationContentDimensions returns the usable area inside Shellforge's
// one-cell border on every side.
func ApplicationContentDimensions(width, height int) (int, int) {
	innerWidth := width - ApplicationBorderSize - 2*ApplicationHorizontalMargin
	innerHeight := height - ApplicationBorderSize
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}
	return innerWidth, innerHeight
}

// WithAppFrame centers content inside a white border that fills the
// outer terminal, keeping the visual frame consistent across all screens.
func WithAppFrame(content string, width, height int) string {
	innerWidth, innerHeight := ApplicationContentDimensions(width, height)
	content = lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, content)
	return ApplicationFrameStyle.Render(content)
}
