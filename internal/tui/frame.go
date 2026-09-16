package tui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	applicationBorderSize       = 2
	applicationHorizontalMargin = 1
)

var applicationFrameStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(whiteColor).
	Padding(0, applicationHorizontalMargin, 0, applicationHorizontalMargin)

// applicationContentDimensions returns the usable area inside Shellforge's
// one-cell border on every side.
func applicationContentDimensions(width, height int) (int, int) {
	innerWidth := width - applicationBorderSize - 2*applicationHorizontalMargin
	innerHeight := height - applicationBorderSize
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}
	return innerWidth, innerHeight
}

// withDisplayApplicationFrame centers content inside a white border that fills the
// outer terminal, keeping the visual frame consistent across all screens.
func withDisplayApplicationFrame(content string, width, height int) string {
	innerWidth, innerHeight := applicationContentDimensions(width, height)
	content = lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, content)
	return applicationFrameStyle.Render(content)
}
