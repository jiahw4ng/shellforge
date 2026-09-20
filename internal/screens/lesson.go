package screens

import (
	"fmt"
	"shellforge/internal/lessonrender"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const lessonPaneGap = 1

// LessonRenderParams holds the parameters for rendering a lesson screen.
type LessonRenderParams struct {
	Lesson           *lessons.Lesson
	Page             *lessons.Page
	PageIndex        int
	TerminalContent  string
	TerminalError    error
	TerminalLoading  string
	TerminalStarting bool
	Width            int
	Height           int
}

// Lesson renders one lesson's instructions beside its embedded sandbox terminal.
func Lesson(p LessonRenderParams) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lessonInstructions(p),
		lessonDivider(p.Height),
		lessonTerminal(p),
	)
}

// LessonTerminalDimensions returns the usable Bubbleterm size for the lesson's
// right-hand pane, including safe defaults before the first resize event.
func LessonTerminalDimensions(width, height int) (int, int) {
	_, terminalWidth := lessonPaneWidths(dimension(width, 80))
	return terminalWidth, dimension(height, 24)
}

// lessonDivider returns a vertical divider string of the given height
func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(ui.WhiteColor).Render(strings.Repeat("│\n", height-1) + "│")
}

// lessonPaneWidths returns the widths of the left and right panes of a lesson, given the total width of the screen.
// the panes will be split evenly, with a gap of 1 character between them
func lessonPaneWidths(width int) (int, int) {
	if width <= lessonPaneGap+2 {
		return 1, 1
	}

	available := width - lessonPaneGap
	rightWidth := available / 2
	return available - rightWidth, rightWidth
}

// lessonInstructions renders the left-hand pane of a lesson, which is the lesson's
// instructions and navigation hints.
func lessonInstructions(p LessonRenderParams) string {
	leftWidth, _ := lessonPaneWidths(p.Width)

	title := fmt.Sprintf("Lesson %d: %s", p.Lesson.Number, p.Lesson.Title)
	pageLabel := fmt.Sprintf("Page %d of %d: %s", p.PageIndex+1, len(p.Lesson.Pages), p.Page.Title)
	markdown, err := lessonrender.Render(p.Page.Content, leftWidth)
	if err != nil {
		markdown = string(p.Page.Content)
	}
	content := strings.Join([]string{
		ui.TitleStyle.Render(title),
		"",
		ui.MutedStyle.Render(pageLabel),
		"",
		markdown,
		"",
		ui.MutedStyle.Render("Ctrl+[ previous page · Ctrl+] next page"),
		ui.MutedStyle.Render("Ctrl+D to return to lesson list"),
	}, "\n")

	return lipgloss.NewStyle().Width(leftWidth).Height(p.Height).Render(content)
}

// lessonTerminal renders the right-hand pane of a lesson, which is either the
// embedded sandbox terminal or
// "starting sandboxed shell..." if the terminal is still initializing or
// a formatted error message
func lessonTerminal(p LessonRenderParams) string {
	_, rightWidth := lessonPaneWidths(p.Width)

	if p.TerminalContent == "" {
		spinner := ""
		if p.TerminalStarting {
			spinner = p.TerminalLoading
		}
		p.TerminalContent = TerminalStart(p.TerminalError, spinner)
	}
	return lipgloss.NewStyle().Width(rightWidth).Height(p.Height).Render(p.TerminalContent)
}

func dimension(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
