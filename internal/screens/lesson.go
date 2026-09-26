package screens

import (
	"fmt"
	"shellforge/internal/assertion"
	"shellforge/internal/lessonrender"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

// LessonRenderParams holds the parameters for rendering a lesson screen.
type LessonRenderParams struct {
	Lesson             *lessons.Lesson
	PageIndex          int
	Guide              viewport.Model
	TerminalContent    string
	TerminalError      error
	TerminalLoading    string
	TerminalStarting   bool
	AssertionResults   []assertion.Result
	AssertionsChecking bool
	AssertionsChecked  bool
	Help               string
	Width              int
	Height             int
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

// PrepareLessonGuide sizes and fills the viewport used for the active page's
// markdown. The lesson title, page label, status, and controls stay fixed while
// the guide itself scrolls between them.
func PrepareLessonGuide(p LessonRenderParams) viewport.Model {
	leftWidth, _ := lessonPaneWidths(p.Width)
	header, footer, markdown := lessonInstructionSections(p, leftWidth)
	height := max(p.Height-lipgloss.Height(header)-lipgloss.Height(footer), 1)

	p.Guide.SetWidth(leftWidth)
	p.Guide.SetHeight(height)
	p.Guide.SetContent(markdown)
	return p.Guide
}

// LessonTerminalDimensions returns the usable Bubbleterm size for the lesson's
// right-hand pane, including safe defaults before the first resize event.
func LessonTerminalDimensions(width, height int) (int, int) {
	_, terminalWidth := lessonPaneWidths(ui.DimensionWithFallback(width, 80))
	return terminalWidth, ui.DimensionWithFallback(height, 24)
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
	p.Guide = PrepareLessonGuide(p)
	header, footer, _ := lessonInstructionSections(p, leftWidth)
	content := header + p.Guide.View() + footer

	return lipgloss.NewStyle().Width(leftWidth).Height(p.Height).Render(content)
}

func lessonInstructionSections(p LessonRenderParams, leftWidth int) (header, footer, markdown string) {
	currPage := p.Lesson.Pages[p.PageIndex]

	title := fmt.Sprintf("Lesson %d: %s", p.Lesson.Number, p.Lesson.Title)
	pageLabel := fmt.Sprintf("Page %d of %d: %s", p.PageIndex+1, len(p.Lesson.Pages), currPage.Title)
	markdown, err := lessonrender.Render(currPage.Content, leftWidth)
	if err != nil {
		markdown = string(currPage.Content)
	}
	header = strings.Join([]string{
		ui.TitleStyle.Render(title),
		"",
		ui.MutedStyle.Render(pageLabel),
		"",
		"",
	}, "\n")

	footerParts := make([]string, 0, 2)
	if status := assertionStatus(p); status != "" {
		footerParts = append(footerParts, status)
	}
	if p.Help != "" {
		footerParts = append(footerParts, p.Help)
	}
	if len(footerParts) > 0 {
		footer = "\n\n" + strings.Join(footerParts, "\n\n")
	}
	return header, footer, markdown
}

// assertionStatus renders the latest progress-check result beneath the lesson
// material without covering the learner's terminal.
func assertionStatus(p LessonRenderParams) string {
	if p.AssertionsChecking {
		return ui.MutedStyle.Render("Checking progress...")
	}
	if !p.AssertionsChecked {
		return ""
	}
	if len(p.AssertionResults) == 0 {
		return "There are no progress checks for this page."
	}

	lines := make([]string, 0, len(p.AssertionResults)+1)
	passed := true
	for _, result := range p.AssertionResults {
		if result.Passed {
			lines = append(lines, ui.SuccessStyle.Render("✓ "+result.Message))
			continue
		}
		passed = false
		lines = append(lines, ui.FailureStyle.Render("✗ "+result.Message))
	}
	if passed && p.Lesson.SuccessMessage != "" {
		lines = append(lines, ui.SuccessStyle.Render(p.Lesson.SuccessMessage))
	} else {
		lines = append(lines, ui.FailureStyle.Render("One or more progress checks did not pass!"))
	}
	return strings.Join(lines, "\n")
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
