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
	"github.com/samber/lo"
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
	wasAtBottom := p.Guide.TotalLineCount() > p.Guide.Height() && p.Guide.AtBottom()
	// Concatenating the header, guide, and footer shares one boundary line per
	// join, so their separately measured heights overlap by two lines.
	height := max(p.Height-lipgloss.Height(header)-lipgloss.Height(footer)+2, 1)

	p.Guide.SetWidth(leftWidth)
	p.Guide.SetHeight(height)
	p.Guide.SetContent(markdown)
	if p.Guide.TotalLineCount() > p.Guide.Height() {
		p.Guide.SetHeight(max(height-1, 1))
		if wasAtBottom || p.Guide.PastBottom() {
			p.Guide.GotoBottom()
		}
	}
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

// lessonDivider renders the vertical rule separating the instruction and
// terminal panes at the requested height.
func lessonDivider(height int) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Foreground(ui.WhiteColor).Render(strings.Repeat("│\n", height-1) + "│")
}

// lessonPaneWidths divides the available width between the instruction and
// terminal panes while reserving one column for the divider.
func lessonPaneWidths(width int) (int, int) {
	if width <= lessonPaneGap+2 {
		return 1, 1
	}

	available := width - lessonPaneGap
	rightWidth := available / 2
	return available - rightWidth, rightWidth
}

// lessonInstructions renders the fixed header and footer around the scrollable
// guide in the lesson's left-hand pane.
func lessonInstructions(p LessonRenderParams) string {
	leftWidth, _ := lessonPaneWidths(p.Width)
	p.Guide = PrepareLessonGuide(p)
	header, footer, _ := lessonInstructionSections(p, leftWidth)
	content := header + p.Guide.View()
	if scrollStatus := lessonGuideScrollStatus(p.Guide); scrollStatus != "" {
		content += "\n" + scrollStatus
	}
	content += footer

	return lipgloss.NewStyle().Width(leftWidth).Height(p.Height).Render(content)
}

// lessonInstructionSections returns the fixed header and footer together with
// the rendered Markdown for the active lesson page.
func lessonInstructionSections(p LessonRenderParams, leftWidth int) (header, footer, markdown string) {
	page := p.Lesson.Pages[p.PageIndex]
	return lessonInstructionHeader(p, page), lessonInstructionFooter(p, leftWidth), lessonInstructionMarkdown(page, leftWidth)
}

// lessonInstructionHeader renders the lesson title, page paginator, and current
// page label above the scrollable guide.
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

// lessonInstructionMarkdown renders the active page to the instruction-pane
// width and translates rendering failures into learner-facing output.
func lessonInstructionMarkdown(page lessons.Page, width int) string {
	markdown, err := lessonrender.Render(page.Content, width)
	if err != nil {
		return ui.FailureStyle.Render(lessonRenderFailureMessage)
	}
	return markdown
}

// lessonInstructionFooter renders hints, progress checks, and help content in
// their display order below the scrollable guide.
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

// lessonHints renders the hints disclosure with the revealed hints and
// the prompt for revealing the next one (if more hints remain)
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
	body = ui.TrimLeadingBlankLines(strings.Trim(body, "\n"))
	if hintCount < len(p.Lesson.Hints) {
		prompt := ui.HintStyle.Render(hintPrompt)
		if body == "" {
			body = prompt
		} else {
			body += "\n" + prompt
		}
	}
	return ui.Disclosure(ui.DisclosureOptions{
		Binding:   config.Keys.ToggleHints,
		Expanded:  p.HintsExpanded,
		Label:     config.HintsDisclosureLabel,
		Body:      body,
		EmptyBody: ui.HintStyle.Render(hintPrompt),
		Width:     width,
	})
}

// lessonPaginator renders one dot per lesson page and highlights the active
// page.
func lessonPaginator(pageIndex, totalPages int) string {
	p := paginator.New()
	p.Type = paginator.Dots
	p.Page = pageIndex
	p.SetTotalPages(totalPages)
	p.ActiveDot = lipgloss.NewStyle().Foreground(ui.AccentColor).Render(paginatorDot)
	p.InactiveDot = ui.MutedStyle.Render(paginatorDot)
	return p.View()
}

// lessonGuideScrollStatus renders muted cues for lesson content above or below
// the viewport, or an empty string when the entire guide is visible.
func lessonGuideScrollStatus(guide viewport.Model) string {
	if guide.TotalLineCount() <= guide.Height() {
		return ""
	}

	var labels []string
	if !guide.AtTop() {
		labels = append(labels, moreAboveLabel)
	}
	if !guide.AtBottom() {
		labels = append(labels, moreBelowLabel)
	}
	return ui.MutedStyle.Render(strings.Join(labels, scrollStatusSeparator))
}

// assertionStatus renders the progress-check disclosure for the current ready,
// checking, completed, or terminal-unavailable state.
func assertionStatus(p LessonRenderParams, width int) string {
	body := ""
	if p.TerminalError != nil {
		// A terminal failure makes progress checks unavailable.
		body = ui.WarningStyle.Render("Progress checks are not available")
	} else if p.AssertionsChecking {
		// Keep every task pending until the current check finishes.
		body = strings.Join([]string{pendingAssertionTasks(p), ui.PrimaryTextStyle.Render(progressCheckingMessage)}, "\n")
	} else if p.AssertionsChecked {
		body = assertionResults(p)
	} else {
		body = strings.Join([]string{pendingAssertionTasks(p), ui.PrimaryTextStyle.Render(progressCheckPrompt)}, "\n")
	}
	return ui.Disclosure(ui.DisclosureOptions{
		Binding:   config.Keys.ToggleAssertions,
		Expanded:  p.AssertionsExpanded,
		Label:     config.ProgressChecksDisclosureLabel,
		Body:      body,
		EmptyBody: ui.PrimaryTextStyle.Render(progressCheckPrompt),
		Width:     width,
	})
}

// pendingAssertionTasks renders every lesson assertion as a pending task.
func pendingAssertionTasks(p LessonRenderParams) string {
	if p.Lesson == nil {
		return ""
	}

	return strings.Join(lo.Map(p.Lesson.Assertions, func(check assertion.Assertion, _ int) string {
		return ui.TextStyle.Render(pendingCheckPrefix + check.Task())
	}), "\n")
}

// assertionResults renders individual check outcomes followed by the appropriate
// aggregate status. An empty result set is reported as a successful no-op.
func assertionResults(p LessonRenderParams) string {
	if len(p.AssertionResults) == 0 {
		return ui.SuccessStyle.Render(passedCheckPrefix + noProgressChecksMessage)
	}

	lines := make([]string, 0, len(p.AssertionResults)+2)
	allPassed := true
	unableToRun := false
	for _, result := range p.AssertionResults {
		if result.Err != nil {
			// Execution errors take precedence over ordinary incomplete results.
			allPassed = false
			unableToRun = true
			message := fmt.Sprintf("Could not check %s: %v", result.Message, result.Err)
			lines = append(lines, ui.FailureStyle.Render(failedCheckPrefix+message))
		} else if result.Passed {
			lines = append(lines, ui.SuccessStyle.Render(passedCheckPrefix+result.Message))
		} else {
			allPassed = false
			lines = append(lines, ui.TextStyle.Render(pendingCheckPrefix+result.Message))
		}
	}
	if unableToRun {
		lines = append(lines, ui.FailureStyle.Render(checksUnavailableMessage))
	} else if allPassed {
		lines = append(lines, ui.SuccessStyle.Render(p.Lesson.SuccessMessage), ui.SuccessStyle.Render(allChecksPassedMessage))
	} else {
		lines = append(lines, ui.PrimaryTextStyle.Render(progressCheckPrompt))
	}
	return strings.Join(lines, "\n")
}

// lessonTerminal renders the lesson's terminal pane, deriving loading or error
// content when no live terminal content is available.
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
