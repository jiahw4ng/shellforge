package screens

import (
	"fmt"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const lessonPaneGap = 1

// Lesson renders one lesson's instructions beside its embedded sandbox terminal.
func Lesson(lesson lessons.Lesson, page lessons.Page, pageIndex int, terminalContent string, terminalError error, width, height int) string {
	leftWidth, rightWidth := lessonPaneWidths(width)
	leftPane := lessonInstructions(lesson, page, pageIndex, leftWidth, height)
	rightPane := lessonTerminal(terminalContent, terminalError, rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, lessonDivider(height), rightPane)
}

// LessonTerminalDimensions returns the usable Bubbleterm size for the lesson's
// right-hand pane, including safe defaults before the first resize event.
func LessonTerminalDimensions(width, height int) (int, int) {
	_, terminalWidth := lessonPaneWidths(dimension(width, 80))
	return terminalWidth, dimension(height, 24)
}

func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(ui.WhiteColor).Render(strings.Repeat("│\n", height-1) + "│")
}

func lessonPaneWidths(width int) (int, int) {
	if width <= lessonPaneGap+2 {
		return 1, 1
	}

	available := width - lessonPaneGap
	rightWidth := available / 2
	return available - rightWidth, rightWidth
}

func lessonInstructions(lesson lessons.Lesson, page lessons.Page, pageIndex, width, height int) string {
	title := fmt.Sprintf("Lesson %d: %s", lesson.Number, lesson.Title)
	pageLabel := fmt.Sprintf("Page %d of %d: %s", pageIndex+1, len(lesson.Pages), page.Title)
	content := strings.Join([]string{
		ui.TitleStyle.Render(title),
		"",
		ui.MutedStyle.Render(pageLabel),
		"",
		page.Content,
		"",
		ui.MutedStyle.Render("The sandbox on the right is ready for this lesson."),
		ui.MutedStyle.Render("Ctrl+< previous page · Ctrl+> next page"),
		ui.MutedStyle.Render("Press Ctrl+D in Bash to return to the lesson list."),
	}, "\n")

	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

func lessonTerminal(terminalContent string, terminalError error, width, height int) string {
	if terminalContent == "" {
		terminalContent = TerminalStart(terminalError)
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(terminalContent)
}

func dimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
