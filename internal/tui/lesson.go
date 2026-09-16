package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const lessonPaneGap = 1

// displayLessonView renders one lesson's instructions beside its embedded
// sandbox terminal, using approximately half the terminal width for each pane.
func displayLessonView(lessonNumber int, terminalContent string, terminalError error, width, height int) string {
	leftWidth, rightWidth := lessonPaneWidths(width)
	leftPane := lessonInstructionsView(lessonNumber, leftWidth, height)
	rightPane := lessonTerminalView(terminalContent, terminalError, rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, lessonDivider(height), rightPane)
}

// lessonDivider fills the gap between the panes with a white vertical rule.
func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(whiteColor).Render(strings.Repeat("│\n", height-1) + "│")
}

// lessonPaneWidths reserves one column between the panes and gives the right
// terminal pane the remaining half of the available width.
func lessonPaneWidths(width int) (int, int) {
	if width <= lessonPaneGap+2 {
		return 1, 1
	}

	available := width - lessonPaneGap
	rightWidth := available / 2
	return available - rightWidth, rightWidth
}

// lessonTerminalDimensions returns the usable Bubbleterm dimensions for the
// right-hand pane, falling back before the first outer resize event arrives.
func lessonTerminalDimensions(width, height int) (int, int) {
	_, terminalWidth := lessonPaneWidths(terminalDimension(width, 80))
	return terminalWidth, terminalDimension(height, 24)
}

// lessonInstructionsView builds the left-hand lesson placeholder content.
func lessonInstructionsView(lessonNumber, width, height int) string {
	title := fmt.Sprintf("Lesson %d: %s", lessonNumber+1, lessonItems[lessonNumber])
	content := strings.Join([]string{
		titleStyle.Render(title),
		"",
		"feature coming soon!",
		"",
		muted.Render("The sandbox on the right is ready for this lesson."),
		muted.Render("Press Ctrl+D in Bash to return to the lesson list."),
	}, "\n")

	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

// lessonTerminalView builds the right-hand pane from Bubbleterm's rendered
// screen, or from its startup/error view before a terminal is ready.
func lessonTerminalView(terminalContent string, terminalError error, width, height int) string {
	if terminalContent == "" {
		terminalContent = terminalStartView(terminalError)
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(terminalContent)
}
