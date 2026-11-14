package app

import (
	"context"
	"log/slog"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/screens"
	"shellforge/internal/storage"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/samber/lo"
)

// handleLessonKeyPress updates lesson progress, sandbox, hints, page, or guide state during a lesson
// forwards all other keys to the embedded terminal.
func (s State) handleLessonKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, config.Keys.ToggleHints):
		s.toggleHints()
		return s, nil
	case key.Matches(msg, config.Keys.ToggleAssertions):
		s.toggleAssertions()
		return s, nil
	case key.Matches(msg, config.Keys.RevealHint):
		s.revealNextHint()
		return s, nil
	case key.Matches(msg, config.Keys.CheckProgress):
		return s, s.startProgressCheck()
	case key.Matches(msg, config.Keys.ResetSandbox):
		return s, s.restartLessonSandbox()
	case key.Matches(msg, config.Keys.PreviousPage):
		s.showPreviousLessonPage()
		return s, nil
	case key.Matches(msg, config.Keys.NextPage):
		s.showNextLessonPage()
		return s, nil
	case key.Matches(msg, config.Keys.GuidePageUp):
		s.scrollLessonGuideUp()
		return s, nil
	case key.Matches(msg, config.Keys.GuidePageDown):
		s.scrollLessonGuideDown()
		return s, nil
	default:
		return s.handleTerminalKeyPress(msg)
	}
}

func (s *State) startProgressCheck() tea.Cmd {
	if s.term.session == nil || s.lessons.progress.isChecking {
		return nil
	}
	s.lessons.progress.isChecking = true
	s.lessons.progress.assertionsExpanded = true
	lesson := s.lessons.available[s.lessons.activeIdx]
	return checkAssertions(s.term.session, lesson.ID, lesson.Assertions, s.term.gen)
}

// checkAssertions evaluates lesson outcomes without blocking the TUI or PTY.
func checkAssertions(session *terminal.TermSession, lessonID string, assertions []assertion.Assertion, generation uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return assertionsCheckedMsg{lessonID: lessonID, results: session.EvaluateAssertions(ctx, assertions), gen: generation}
	}
}

// lessonTerminalContent is what the lesson pane passes down for its terminal:
// the live terminal or the reason it is not running. It is empty while the
// terminal starts, so the screen can show its own loading state.
func (s State) lessonTerminalContent() string {
	if s.term.session != nil {
		return s.term.session.View()
	}
	content, _ := s.terminalStoppedView()
	return content
}

func (s *State) restartLessonSandbox() tea.Cmd {
	if s.term.isStarting {
		return nil
	}
	s.CloseTerminal()
	return s.startTerminalWithActiveLesson()
}

func (s *State) revealNextHint() {
	hints := s.lessons.available[s.lessons.activeIdx].Hints
	if s.lessons.revealedHints >= len(hints) {
		return
	}
	s.lessons.revealedHints++
	s.lessons.hintsExpanded = true
	s.prepareLessonGuide()
}

func (s *State) toggleHints() {
	s.lessons.hintsExpanded = !s.lessons.hintsExpanded
	s.prepareLessonGuide()
}

func (s *State) toggleAssertions() {
	s.lessons.progress.assertionsExpanded = !s.lessons.progress.assertionsExpanded
	s.prepareLessonGuide()
}

func (s *State) scrollLessonGuideUp() {
	s.prepareLessonGuide()
	s.lessons.guide.ScrollUp(1)
}

func (s *State) scrollLessonGuideDown() {
	s.prepareLessonGuide()
	s.lessons.guide.ScrollDown(1)
}

func (s *State) showPreviousLessonPage() {
	if s.lessons.activePage > 0 {
		s.showLessonPage(s.lessons.activePage - 1)
	}
}

func (s *State) showNextLessonPage() {
	pages := s.lessons.available[s.lessons.activeIdx].Pages
	if s.lessons.activePage < len(pages)-1 {
		s.showLessonPage(s.lessons.activePage + 1)
	}
}

// showLessonPage turns to a page, collapsing help and scrolling back to the top.
func (s *State) showLessonPage(page int) {
	s.lessons.activePage = page
	s.help.ShowAll = false
	s.lessons.guide.GotoTop()
}

func (s State) lessonView() string {
	return s.lessonScreen().View()
}

// assertionsCheckedMsg carries the results of an asynchronous lesson check.
type assertionsCheckedMsg struct {
	lessonID string
	results  []assertion.Result
	gen      uint64
}

type completionSavedMsg struct {
	lessonID string
	err      error
}

// startTerminalWithActiveLesson creates a fresh sandbox for the lesson currently
// displayed beside the terminal.
func (s *State) startTerminalWithActiveLesson() tea.Cmd {
	lesson := &s.lessons.available[s.lessons.activeIdx]
	return s.startTerminal(lesson)
}

func (s State) lessonScreen() screens.Lesson {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	instructionWidth := screens.LessonInstructionWidth(ui.DimensionWithFallback(width, 80))
	lesson := s.lessons.available[s.lessons.activeIdx]
	pageIndex := s.lessons.activePage
	return screens.Lesson{
		Lesson:             &lesson,
		PageIndex:          pageIndex,
		Guide:              s.lessons.guide,
		RevealedHints:      s.lessons.revealedHints,
		HintsExpanded:      s.lessons.hintsExpanded,
		TerminalContent:    s.lessonTerminalContent(),
		TerminalError:      s.term.err,
		TerminalLoading:    s.term.spinner.View(),
		TerminalSteps:      s.term.startSteps,
		TerminalStarting:   s.term.isStarting,
		AssertionResults:   s.lessons.progress.results,
		AssertionsChecking: s.lessons.progress.isChecking,
		AssertionsChecked:  s.lessons.progress.hasChecked,
		AssertionsExpanded: s.lessons.progress.assertionsExpanded,
		Help:               s.lessonHelp(instructionWidth),
		Width:              width,
		Height:             height,
	}
}

func (s *State) prepareLessonGuide() {
	if s.nav.screen != config.LessonTerminalScreen {
		return
	}
	s.lessons.guide = s.lessonScreen().PrepareGuide()
}

// handleAssertionsChecked stores current-generation assertion results and marks
// a newly passing lesson complete, scheduling persistence when available.
func (s State) handleAssertionsChecked(msg assertionsCheckedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != s.term.gen {
		return s, nil
	}
	s.recordAssertionResults(msg.results)

	hasPassedAllAssertions := lo.EveryBy(msg.results, func(result assertion.Result) bool {
		return result.Passed
	})
	if msg.lessonID == "" || !hasPassedAllAssertions {
		return s, nil
	}
	return s, s.completeLesson(msg.lessonID)
}

// recordAssertionResults stores the latest check outcome and shows it.
func (s *State) recordAssertionResults(results []assertion.Result) {
	s.lessons.progress.isChecking = false
	s.lessons.progress.hasChecked = true
	s.lessons.progress.results = results
	s.lessons.progress.assertionsExpanded = true
}

// completeLesson marks a lesson complete the first time it passes. It returns
// the command that persists the completion, or nil when the lesson was already
// complete or storage is unavailable.
func (s *State) completeLesson(lessonID string) tea.Cmd {
	if s.lessons.completed[lessonID] {
		return nil
	}
	s.lessons.completed[lessonID] = true
	if s.lessons.store == nil {
		return nil
	}
	return saveCompletion(s.lessons.store, lessonID)
}

// handleCompletionSaved logs any persistence error and flags it for the help
// notice. The lesson was already marked complete in memory, so completion
// itself is unchanged.
func (s State) handleCompletionSaved(msg completionSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		slog.Error("Could not persist lesson completion", "lesson_id", msg.lessonID, logErrorKey, msg.err)
		s.lessons.progressNotSaved = true
	}
	return s, nil
}

func saveCompletion(store storage.CompletionStore, lessonID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := store.MarkLessonIDAsCompleted(ctx, lessonID)
		return completionSavedMsg{lessonID: lessonID, err: err}
	}
}

// lessonHelp renders toggleable shortcuts for the current lesson state.
func (s State) lessonHelp(width int) string {
	return s.renderHelp(width, config.LessonHelpKeyBindingGroups)
}

type progressState struct {
	// results contains the outcomes of the most recent assertion check.
	results []assertion.Result
	// isChecking reports whether an assertion command is currently running.
	isChecking bool
	// hasChecked distinguishes no check yet from a completed check with no results.
	hasChecked bool
	// assertionsExpanded controls whether completed results are shown in the lesson footer.
	assertionsExpanded bool
}
