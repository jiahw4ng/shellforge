package tui

import (
	"fmt"
	lessondefs "shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const lessonPaneGap = 1

// displayLessonView renders one lesson's instructions beside its embedded
// sandbox terminal, using approximately half the terminal width for each pane.
func displayLessonView(lesson lessondefs.Lesson, terminalContent string, terminalError error, width, height int) string {
	leftWidth, rightWidth := lessonPaneWidths(width)
	leftPane := lessonInstructionsView(lesson, leftWidth, height)
	rightPane := lessonTerminalView(terminalContent, terminalError, rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, lessonDivider(height), rightPane)
}

// lessonDivider fills the gap between the panes with a white vertical rule.
func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(ui.WhiteColor).Render(strings.Repeat("│\n", height-1) + "│")
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

// lessonInstructionsView builds the left-hand lesson instructions from its YAML definition.
func lessonInstructionsView(lesson lessondefs.Lesson, width, height int) string {
	title := fmt.Sprintf("Lesson %d: %s", lesson.Number, lesson.Title)
	content := strings.Join([]string{
		ui.TitleStyle.Render(title),
		"",
		lesson.Content,
		"",
		ui.MutedStyle.Render("The sandbox on the right is ready for this lesson."),
		ui.MutedStyle.Render("Press Ctrl+D in Bash to return to the lesson list."),
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
