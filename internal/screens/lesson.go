package screens

import (
	"fmt"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/lessonrender"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"

	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// LessonRenderParams holds the parameters for rendering a lesson screen.
type LessonRenderParams struct {
	Lesson             *lessons.Lesson
	PageIndex          int
	Guide              viewport.Model
	RevealedHints      int
	HintsExpanded      bool
	TerminalContent    string
	TerminalError      error
	TerminalLoading    string
	TerminalStarting   bool
	AssertionResults   []assertion.Result
	AssertionsChecking bool
	AssertionsChecked  bool
	AssertionsExpanded bool
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
	// Concatenating the header, guide, and footer shares one boundary line per
	// join, so their separately measured heights overlap by two lines.
	height := max(p.Height-lipgloss.Height(header)-lipgloss.Height(footer)+2, 1)

	p.Guide.SetWidth(leftWidth)
	p.Guide.SetHeight(height)
	p.Guide.SetContent(markdown)
	return p.Guide
}

// LessonTerminalDimensions returns the usable Bubbleterm size for the lesson's
// right-hand pane, including safe defaults before the first resize event.
func LessonTerminalDimensions(width, height int) (int, int) {
	_, terminalWidth := lessonPaneWidths(ui.DimensionWithFallback(width, 80))
	return terminalWidth, TerminalContentHeight(height)
}

// LessonInstructionWidth returns the width of the lesson instruction pane.
func LessonInstructionWidth(width int) int {
	leftWidth, _ := lessonPaneWidths(ui.DimensionWithFallback(width, 80))
	return leftWidth
}

// lessonDivider returns a vertical divider with at least one row.
func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(ui.WhiteColor).Render(strings.Repeat("│\n", height-1) + "│")
}

// lessonPaneWidths divides the available width evenly around the fixed pane gap.
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
	page := p.Lesson.Pages[p.PageIndex]
	return lessonInstructionHeader(p, page), lessonInstructionFooter(p, leftWidth), lessonInstructionMarkdown(page, leftWidth)
}

func lessonInstructionHeader(p LessonRenderParams, page lessons.Page) string {
	title := fmt.Sprintf("Lesson %d: %s", p.Lesson.Number, p.Lesson.Title)
	pageLabel := fmt.Sprintf("Page %d of %d: %s", p.PageIndex+1, len(p.Lesson.Pages), page.Title)
	return strings.Join([]string{
		ui.TitleStyle.Render(title),
		lessonPaginator(p.PageIndex, len(p.Lesson.Pages)),
		ui.MutedStyle.Render(pageLabel),
		"",
		"",
	}, "\n")
}

func lessonInstructionMarkdown(page lessons.Page, width int) string {
	markdown, err := lessonrender.Render(page.Content, width)
	if err != nil {
		return string(page.Content)
	}
	return markdown
}

func lessonInstructionFooter(p LessonRenderParams, width int) string {
	footerParts := make([]string, 0, 3)
	if hints := lessonHints(p, width); hints != "" {
		footerParts = append(footerParts, hints)
	}
	if status := assertionStatus(p, width); status != "" {
		footerParts = append(footerParts, status)
	}
	if p.Help != "" {
		footerParts = append(footerParts, p.Help)
	}
	if len(footerParts) > 0 {
		return "\n\n" + strings.Join(footerParts, "\n")
	}
	return ""
}

func lessonHints(p LessonRenderParams, width int) string {
	hintCount := min(p.RevealedHints, len(p.Lesson.Hints))
	var content strings.Builder
	for i, hint := range p.Lesson.Hints[:hintCount] {
		fmt.Fprintf(&content, "%d. %s\n", i+1, hint)
	}
	body, err := lessonrender.RenderHint(lessons.Markdown(content.String()), max(width-2, 1))
	if err != nil {
		body = content.String()
	}
	body = trimLeadingBlankLines(strings.Trim(body, "\n"))
	if hintCount < len(p.Lesson.Hints) {
		prompt := ui.HintStyle.Render("Press F10 to show a hint.")
		if body == "" {
			body = prompt
		} else {
			body += "\n" + prompt
		}
	}
	return ui.Disclosure(
		config.Keys.ToggleHints,
		p.HintsExpanded,
		"show hints",
		"hide hints",
		body,
		ui.HintStyle.Render("Press F10 to show a hint."),
		width,
	)
}

func trimLeadingBlankLines(text string) string {
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && strings.TrimSpace(ansi.Strip(lines[0])) == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

func lessonPaginator(pageIndex, totalPages int) string {
	p := paginator.New()
	p.Type = paginator.Dots
	p.Page = pageIndex
	p.SetTotalPages(totalPages)
	p.ActiveDot = lipgloss.NewStyle().Foreground(ui.BlueColor).Render("•")
	p.InactiveDot = ui.MutedStyle.Render("•")
	return p.View()
}

// assertionStatus renders the latest progress-check result beneath the lesson
// material without covering the learner's terminal.
func assertionStatus(p LessonRenderParams, width int) string {
	body := ""
	if p.AssertionsChecking {
		body = ui.MutedStyle.Render("Checking progress...")
	} else if p.AssertionsChecked {
		body = assertionResults(p)
	}
	return ui.Disclosure(
		config.Keys.ToggleAssertions,
		p.AssertionsExpanded,
		"show assertions",
		"hide assertions",
		body,
		ui.MutedStyle.Render("Press F12 to run assertions."),
		width,
	)
}

func assertionResults(p LessonRenderParams) string {
	if len(p.AssertionResults) == 0 {
		return ui.SuccessStyle.Render("✓ There are no progress checks for this page.")
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
	if passed {
		lines = append(lines, ui.SuccessStyle.Render(p.Lesson.SuccessMessage))
	} else {
		lines = append(lines, ui.FailureStyle.Render("One or more progress checks did not pass!"))
	}
	return strings.Join(lines, "\n")
}

// lessonTerminal renders the embedded sandbox, its startup state, or its
// recoverable error in the lesson's right-hand pane.
func lessonTerminal(p LessonRenderParams) string {
	_, rightWidth := lessonPaneWidths(p.Width)

	if p.TerminalContent == "" {
		spinner := ""
		if p.TerminalStarting {
			spinner = p.TerminalLoading
		}
		p.TerminalContent = TerminalStart(p.TerminalError, spinner)
	}
	return TerminalPane(p.TerminalContent, rightWidth, p.Height)
}
