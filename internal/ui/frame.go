// Package ui provides shared Lipgloss styles and the application frame.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
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

// WithAppFrameJunction renders content inside the application frame and joins
// a content-local vertical divider to the top and bottom borders.
func WithAppFrameJunction(content string, width, height, contentColumn int) string {
	innerWidth, _ := ApplicationContentDimensions(width, height)
	framed := WithAppFrame(content, width, height)
	if contentColumn < 0 || contentColumn >= innerWidth {
		return framed
	}

	lines := strings.Split(framed, "\n")
	if len(lines) < 2 {
		return framed
	}

	frameColumn := ApplicationBorderSize/2 + ApplicationHorizontalMargin + contentColumn
	border := lipgloss.NormalBorder()
	lines[0] = frameBorderWithJunction(lines[0], frameColumn, border.TopLeft, border.Top, border.MiddleTop, border.TopRight)
	last := len(lines) - 1
	lines[last] = frameBorderWithJunction(lines[last], frameColumn, border.BottomLeft, border.Bottom, border.MiddleBottom, border.BottomRight)
	return strings.Join(lines, "\n")
}

func frameBorderWithJunction(line string, column int, left, horizontal, junction, right string) string {
	width := lipgloss.Width(line)
	if column <= 0 || column >= width-1 {
		return line
	}

	border := left +
		strings.Repeat(horizontal, column-1) +
		junction +
		strings.Repeat(horizontal, width-column-2) +
		right
	return lipgloss.NewStyle().Foreground(WhiteColor).Render(border)
}
