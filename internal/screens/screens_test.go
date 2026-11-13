package screens

import (
	"errors"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestMainMenuShowsItems(t *testing.T) {
	items := []string{"Sandbox", "Choose lesson", "Settings", "Exit"}
	view := ansi.Strip(MainMenu(items, 0, "navigation help"))
	for _, text := range append([]string{"Welcome to Shellforge!"}, items...) {
		if !strings.Contains(view, text) {
			t.Errorf("main menu does not contain %q", text)
		}
	}
}

func TestLessonListShowsLoadedLessonsAndBack(t *testing.T) {
	available := []lessons.Lesson{
		{ID: "01-navigation", Number: 1, Title: "Getting around", Description: "Explore a project."},
		{ID: "02-files", Number: 2, Title: "Files and directories", Description: "Create and organize files."},
	}
	view := LessonList(available, map[string]bool{"01-navigation": true}, 1, nil, "navigation help")
	for _, text := range []string{"1. Getting around", "Back"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson list does not contain %q", text)
		}
	}
	plainView := ansi.Strip(view)
	if !strings.Contains(plainView, "✓ 1. Getting around") {
		t.Error("lesson list does not mark the completed lesson")
	}
	if strings.Contains(plainView, "✓ 2. Files and directories") {
		t.Error("lesson list marks an incomplete lesson")
	}
	if !strings.Contains(plainView, "     └── Create and organize files.") {
		t.Error("lesson list does not contain the selected lesson's indented description")
	}
	if strings.Contains(view, "Explore a project.") {
		t.Error("lesson list contains an unselected lesson's description")
	}
}

func TestLessonUsesHalfWidthTerminal(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	view := Lesson{
		Lesson:          &lesson,
		PageIndex:       0,
		TerminalContent: "shellforge$ ",
		TerminalError:   nil,
		Width:           100,
		Height:          24,
	}.View()

	plainView := ansi.Strip(view)
	for _, text := range []string{"Lesson 1: Getting around", "Page 1 of 1: pwd", "Your task", ">_ Sandboxed Bash", "shellforge$ ", "│"} {
		if !strings.Contains(plainView, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}

	left, right := lessonPaneWidths(100)
	if left+right+lessonPaneGap != 100 || right != 49 {
		t.Fatalf("lesson pane widths = %d and %d, want 50 and 49", left, right)
	}
}

func TestSandboxTerminalPaneTitleNamesTheKeyThatLeaves(t *testing.T) {
	plainView := ansi.Strip(SandboxTerminalPane("shellforge$ ", 80, 24))
	lines := strings.Split(plainView, "\n")

	want := ">_ Sandboxed Bash (type exit to quit)"
	if strings.TrimSpace(lines[0]) != want {
		t.Errorf("sandbox pane title = %q, want %q", strings.TrimSpace(lines[0]), want)
	}
	if !strings.Contains(plainView, "shellforge$ ") {
		t.Error("sandbox pane does not contain terminal content")
	}
}

func TestLessonTerminalPaneTitleHasNoKeyHint(t *testing.T) {
	plainView := ansi.Strip(TerminalPane("shellforge$ ", 80, 24))

	if strings.Contains(plainView, "press") || strings.Contains(plainView, "to quit") {
		t.Errorf("lesson terminal pane title carries a key hint: %q", plainView)
	}
}

func TestTerminalPaneShowsHeaderAboveContent(t *testing.T) {
	view := TerminalPane("shellforge$ ", 80, 24)
	plainView := ansi.Strip(view)
	lines := strings.Split(plainView, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != ">_ Sandboxed Bash" {
		t.Fatalf("terminal pane does not place its prompt-styled header above the terminal: %q", plainView)
	}
	if !strings.HasPrefix(view, ui.ControlStyle.Render(sandboxTitle)) {
		t.Fatalf("terminal pane header does not use the control style: %q", view)
	}
	if !strings.Contains(plainView, "shellforge$ ") {
		t.Fatal("terminal pane does not contain terminal content")
	}
	if height := TerminalContentHeight(24); height != 23 {
		t.Fatalf("terminal content height = %d, want 23", height)
	}
}

func TestLessonPageLabelHasSpacingBeforeGuide(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	header, _, _ := lessonInstructionSections(Lesson{
		Lesson:    &lesson,
		PageIndex: 0,
	}, 50)

	if !strings.HasSuffix(ansi.Strip(header), "Page 1 of 1: pwd\n\n") {
		t.Fatalf("lesson header = %q, want a blank line after the page label", ansi.Strip(header))
	}
}

func TestLessonHeaderShowsPaginatorDotsBetweenTitleAndPageLabel(t *testing.T) {
	lesson := lessons.Lesson{
		Number: 1,
		Title:  "Getting around",
		Pages: []lessons.Page{
			{Title: "first"},
			{Title: "second"},
			{Title: "third"},
		},
	}
	header, _, _ := lessonInstructionSections(Lesson{
		Lesson:    &lesson,
		PageIndex: 1,
	}, 50)

	plainHeader := ansi.Strip(header)
	if !strings.HasPrefix(plainHeader, "Lesson 1: Getting around\n•••\nPage 2 of 3: second\n") {
		t.Fatalf("lesson header = %q, want title, paginator dots, then page label", plainHeader)
	}

	wantDots := ui.MutedStyle.Render("•") + lipgloss.NewStyle().Foreground(ui.AccentColor).Render("•") + ui.MutedStyle.Render("•")
	if !strings.Contains(header, wantDots) {
		t.Fatalf("lesson paginator = %q, want only page 2 highlighted", header)
	}
}

func TestLessonKeepsGuideChromeOutsideScrollableContent(t *testing.T) {
	page := lessons.Page{Title: "long guide", Content: lessons.Markdown(strings.Repeat("guide line\n\n", 40))}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	p := Lesson{
		Lesson: &lesson,
		Help:   "PgUp scroll guide up · PgDn scroll guide down",
		Width:  100,
		Height: 24,
	}
	p.Guide = p.PrepareGuide()
	p.Guide.PageDown()

	view := ansi.Strip(p.View())
	for _, text := range []string{"Lesson 1: Getting around", "Page 1 of 1: long guide", "↑ More above", "↓ More below", "PgUp scroll guide up", "PgDn scroll guide down"} {
		if !strings.Contains(view, text) {
			t.Errorf("scrolled lesson view does not contain fixed text %q", text)
		}
	}
}

func TestLessonGuideScrollStatusTracksAvailableContent(t *testing.T) {
	page := lessons.Page{Title: "long guide", Content: lessons.Markdown(strings.Repeat("guide line\n\n", 40))}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	params := Lesson{Lesson: &lesson, Width: 100, Height: 24}

	params.Guide = params.PrepareGuide()
	if view := ansi.Strip(params.View()); !strings.Contains(view, "↓ More below") || strings.Contains(view, "↑ More above") {
		t.Fatalf("guide at top has incorrect scroll status: %q", view)
	}

	params.Guide.GotoBottom()
	if view := ansi.Strip(params.View()); !strings.Contains(view, "↑ More above") || strings.Contains(view, "↓ More below") {
		t.Fatalf("guide at bottom has incorrect scroll status: %q", view)
	}

	shortLesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{{Title: "short", Content: "One line."}}}
	shortParams := Lesson{Lesson: &shortLesson, Width: 100, Height: 24}
	if view := ansi.Strip(shortParams.View()); strings.Contains(view, "More above") || strings.Contains(view, "More below") {
		t.Fatalf("short guide unexpectedly has scroll status: %q", view)
	}
}

func TestLessonShowsOnlyRevealedHintsInDisclosure(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Run a command."}
	lesson := lessons.Lesson{
		Number: 1,
		Title:  "Getting around",
		Pages:  []lessons.Page{page},
		Hints:  []string{"Try `pwd`.", "Then inspect the output."},
	}

	before := ansi.Strip(Lesson{Lesson: &lesson, Width: 100, Height: 24}.View())
	for _, text := range []string{"▶ F1 Hints", "▶ F2 Progress checks"} {
		if !strings.Contains(before, text) {
			t.Fatalf("lesson does not display %q before actions: %q", text, before)
		}
	}
	if strings.Contains(before, "Try pwd") {
		t.Fatalf("lesson displays hint content before F10: %q", before)
	}

	collapsed := ansi.Strip(Lesson{
		Lesson:        &lesson,
		RevealedHints: 1,
		Width:         100,
		Height:        24,
	}.View())
	if !strings.Contains(collapsed, "▶ F1 Hints") || strings.Contains(collapsed, "Try pwd") {
		t.Fatalf("collapsed hint disclosure = %q", collapsed)
	}

	expanded := ansi.Strip(Lesson{
		Lesson:        &lesson,
		RevealedHints: 1,
		HintsExpanded: true,
		Width:         100,
		Height:        24,
	}.View())
	if !strings.Contains(expanded, "▼ F1 Hints") || !strings.Contains(expanded, "Try") || !strings.Contains(expanded, "pwd") {
		t.Fatalf("lesson does not display the first revealed hint: %q", expanded)
	}
	hintSection := ansi.Strip(lessonHints(Lesson{
		Lesson:        &lesson,
		RevealedHints: 1,
		HintsExpanded: true,
	}, 49))
	hintLines := strings.Split(hintSection, "\n")
	if len(hintLines) < 2 || strings.TrimSpace(hintLines[1]) == "" {
		t.Fatalf("expanded hint disclosure starts with a blank body line: %q", hintSection)
	}
	renderedHintSection := lessonHints(Lesson{
		Lesson:        &lesson,
		RevealedHints: 1,
		HintsExpanded: true,
	}, 49)
	if !strings.Contains(renderedHintSection, ui.ControlStyle.Render("▼ F1 Hints")) {
		t.Fatalf("expanded hint disclosure does not use the disclosure style: %q", renderedHintSection)
	}
	if strings.Contains(expanded, "Then inspect the output") {
		t.Fatalf("lesson displays an unrevealed hint: %q", expanded)
	}
	if !strings.Contains(expanded, "└ Press F10 to show a hint.") {
		t.Fatalf("lesson does not prompt for the next available hint: %q", expanded)
	}

	allRevealed := ansi.Strip(lessonHints(Lesson{
		Lesson:        &lesson,
		RevealedHints: len(lesson.Hints),
		HintsExpanded: true,
	}, 49))
	if strings.Contains(allRevealed, "Press F10 to show a hint.") {
		t.Fatalf("lesson prompts for a hint after all hints are revealed: %q", allRevealed)
	}
}

func TestLessonDisclosuresShowPendingInstructions(t *testing.T) {
	lesson := lessons.Lesson{
		Number: 1,
		Title:  "Getting around",
		Pages:  []lessons.Page{{Title: "pwd", Content: "Your task"}},
		Hints:  []string{"Try `pwd`."},
		Assertions: []assertion.Assertion{
			{Type: assertion.AssertionTypeFileExists, Path: "debug/bar.txt"},
		},
	}
	view := ansi.Strip(Lesson{
		Lesson:             &lesson,
		HintsExpanded:      true,
		AssertionsExpanded: true,
		Width:              100,
		Height:             24,
	}.View())
	for _, text := range []string{
		"▼ F1 Hints",
		"└ Press F10 to show a hint.",
		"▼ F2 Progress checks",
		"│ ○ Create file debug/bar.txt",
		"└ Press F12 to check progress.",
	} {
		if !strings.Contains(view, text) {
			t.Errorf("empty lesson disclosures do not contain %q: %q", text, view)
		}
	}
}

func TestLessonAssertionStatusMessagesUseStatusStyles(t *testing.T) {
	tests := []struct {
		name      string
		params    Lesson
		want      string
		style     lipgloss.Style
		doNotWant string
	}{
		{
			name:   "ready",
			params: Lesson{AssertionsExpanded: true},
			want:   "Press F12 to check progress.",
			style:  ui.PrimaryTextStyle,
		},
		{
			name: "checking",
			params: Lesson{
				Lesson: &lessons.Lesson{Assertions: []assertion.Assertion{
					{Type: assertion.AssertionTypeFileExists, Path: "debug/bar.txt"},
				}},
				AssertionsChecking: true,
				AssertionsExpanded: true,
			},
			want:  "Checking Progress...",
			style: ui.PrimaryTextStyle,
		},
		{
			name:      "sandbox unavailable",
			params:    Lesson{TerminalError: errors.New("Docker is unavailable"), AssertionsExpanded: true},
			want:      "Progress checks are not available",
			style:     ui.WarningStyle,
			doNotWant: "Press F12 to check progress.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := assertionStatus(test.params, 100)
			if !strings.Contains(view, test.style.Render(test.want)) {
				t.Errorf("assertion status does not render %q with the expected style: %q", test.want, view)
			}
			if test.name != "sandbox unavailable" && !strings.HasSuffix(ansi.Strip(view), "└ "+test.want) {
				t.Errorf("assertion status does not place %q at the bottom: %q", test.want, ansi.Strip(view))
			}
			if test.doNotWant != "" && strings.Contains(ansi.Strip(view), test.doNotWant) {
				t.Errorf("assertion status contains unavailable action %q: %q", test.doNotWant, view)
			}
		})
	}
}

func TestLessonShowsAssertionResults(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	collapsed := ansi.Strip(Lesson{
		Lesson:             &lesson,
		AssertionResults:   []assertion.Result{{Passed: false, Message: "File does not exist."}},
		AssertionsChecked:  true,
		AssertionsExpanded: false,
		Width:              100,
		Height:             24,
	}.View())
	if !strings.Contains(collapsed, "▶ F2 Progress checks") || strings.Contains(collapsed, "File does not exist.") {
		t.Fatalf("collapsed assertion disclosure = %q", collapsed)
	}

	params := Lesson{
		Lesson: &lesson,
		AssertionResults: []assertion.Result{
			{Passed: true, Message: "Directory exists."},
			{Passed: false, Message: "File does not exist."},
		},
		AssertionsChecked:  true,
		AssertionsExpanded: true,
		Help:               "Ctrl+Alt+R reset sandbox · F12 check progress",
		Width:              100,
		Height:             24,
	}
	view := params.View()
	plainView := ansi.Strip(view)
	for _, text := range []string{"Directory exists.", "File does not exist.", "Ctrl+Alt+R reset sandbox", "F12 check progress"} {
		if !strings.Contains(plainView, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}
	if !strings.Contains(view, ui.TextStyle.Render("○ File does not exist.")) {
		t.Fatalf("incomplete task does not use neutral pending style: %q", view)
	}
	if strings.Contains(view, ui.FailureStyle.Render("✗ File does not exist.")) {
		t.Fatalf("incomplete task uses failure styling: %q", view)
	}
	if strings.Contains(plainView, "One or more progress checks did not pass!") {
		t.Fatalf("incomplete task includes an error summary: %q", plainView)
	}
	statusView := assertionStatus(params, 100)
	wantPrompt := ui.PrimaryTextStyle.Render("Press F12 to check progress.")
	if !strings.Contains(statusView, wantPrompt) || !strings.HasSuffix(ansi.Strip(statusView), "└ Press F12 to check progress.") {
		t.Fatalf("incomplete checks do not restore the white progress prompt at the bottom: %q", statusView)
	}
}

func TestLessonUsesFailureStyleOnlyForProgressCheckErrors(t *testing.T) {
	lesson := lessons.Lesson{Pages: []lessons.Page{{Title: "pwd", Content: "Your task"}}}
	checkErr := errors.New("sandbox stopped")
	view := assertionStatus(Lesson{
		Lesson: &lesson,
		AssertionResults: []assertion.Result{
			{Message: "Create directory debug"},
			{Message: "Create file debug/bar.txt", Err: checkErr},
		},
		AssertionsChecked:  true,
		AssertionsExpanded: true,
	}, 100)

	want := ui.FailureStyle.Render("✗ Could not check Create file debug/bar.txt: sandbox stopped")
	if !strings.Contains(view, want) {
		t.Fatalf("progress-check error does not use failure style: %q", view)
	}
	wantSummary := ui.FailureStyle.Render("Unable to run some checks!")
	if !strings.Contains(view, wantSummary) || !strings.HasSuffix(ansi.Strip(view), "└ Unable to run some checks!") {
		t.Fatalf("progress-check error does not show the failure summary at the bottom: %q", view)
	}
}

func TestLessonFooterOrdersHintsAssertionsThenHelp(t *testing.T) {
	lesson := lessons.Lesson{
		Number: 1,
		Title:  "Getting around",
		Pages:  []lessons.Page{{Title: "pwd", Content: "Your task"}},
		Hints:  []string{"Try `pwd`."},
	}
	view := ansi.Strip(Lesson{
		Lesson:             &lesson,
		RevealedHints:      1,
		HintsExpanded:      true,
		AssertionResults:   []assertion.Result{{Passed: false, Message: "File does not exist."}},
		AssertionsChecked:  true,
		AssertionsExpanded: true,
		Help:               "▶ F3 Help",
		Width:              100,
		Height:             30,
	}.View())

	hintsIndex := strings.Index(view, "▼ F1 Hints")
	assertionsIndex := strings.Index(view, "▼ F2 Progress checks")
	helpIndex := strings.Index(view, "▶ F3 Help")
	if hintsIndex < 0 || assertionsIndex < 0 || helpIndex < 0 {
		t.Fatalf("lesson footer is missing a disclosure: %q", view)
	}
	if hintsIndex >= assertionsIndex || assertionsIndex >= helpIndex {
		t.Fatalf("lesson footer order = hints %d, assertions %d, help %d", hintsIndex, assertionsIndex, helpIndex)
	}
}

func TestLessonFooterDoesNotAddBlankLinesBetweenPanels(t *testing.T) {
	lesson := lessons.Lesson{
		Pages: []lessons.Page{{}},
		Hints: []string{"A hint."},
	}
	footer := ansi.Strip(lessonInstructionFooter(Lesson{
		Lesson:            &lesson,
		RevealedHints:     1,
		AssertionsChecked: true,
		Help:              "▶ F3 Help",
	}, 40))

	want := "▶ F1 Hints\n▶ F2 Progress checks\n▶ F3 Help"
	if strings.TrimSpace(footer) != want {
		t.Fatalf("lesson footer = %q, want adjacent panels %q", footer, want)
	}
}

func TestLessonFooterReachesBottomOfInstructionPane(t *testing.T) {
	lesson := lessons.Lesson{
		Number: 1,
		Title:  "Getting around",
		Pages:  []lessons.Page{{Title: "pwd", Content: "Your task"}},
	}
	view := ansi.Strip(lessonInstructions(Lesson{
		Lesson: &lesson,
		Help:   "▶ F3 Help",
		Width:  100,
		Height: 24,
	}))
	lines := strings.Split(view, "\n")
	if len(lines) != 24 {
		t.Fatalf("instruction pane height = %d, want 24", len(lines))
	}
	if !strings.Contains(lines[len(lines)-1], "▶ F3 Help") {
		t.Fatalf("instruction pane ends with %q, want footer on final row", lines[len(lines)-1])
	}
}

func TestLessonShowsAllChecksPassedForPassingAssertions(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{
		Number:         1,
		Title:          "Getting around",
		Pages:          []lessons.Page{page},
		SuccessMessage: "Navigation complete!",
	}
	view := ansi.Strip(assertionStatus(Lesson{
		Lesson: &lesson,
		AssertionResults: []assertion.Result{
			{Passed: true, Message: "Directory exists."},
		},
		AssertionsChecked:  true,
		AssertionsExpanded: true,
	}, 100))

	if !strings.Contains(view, "All checks passed!") {
		t.Fatalf("lesson does not show the all-checks-passed status: %q", view)
	}
	if !strings.HasSuffix(view, "└ All checks passed!") {
		t.Fatalf("lesson does not place the all-checks-passed status at the bottom: %q", view)
	}
	if !strings.Contains(view, "Navigation complete!") {
		t.Fatalf("lesson omits its success message before the final status: %q", view)
	}
}

func TestLessonHidesProgressStatusUntilChecked(t *testing.T) {
	page := lessons.Page{Title: "Introduction", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}

	beforeCheck := ansi.Strip(Lesson{
		Lesson: &lesson,
		Width:  100,
		Height: 24,
	}.View())
	if strings.Contains(beforeCheck, "There are no progress checks") {
		t.Error("lesson shows the no-progress-checks status before F12")
	}

	afterCheck := Lesson{
		Lesson:             &lesson,
		AssertionsChecked:  true,
		AssertionsExpanded: true,
		Width:              100,
		Height:             24,
	}.View()

	wantStatus := ui.SuccessStyle.Render("✓ There are no progress checks for this page.")
	if !strings.Contains(afterCheck, wantStatus) {
		t.Error("lesson does not show the no-progress-checks status after F12")
	}
}

func TestSettingsShowsResetAndResult(t *testing.T) {
	view := ansi.Strip(Settings([]string{"Reset lesson progress", "Back"}, 0, "Lesson progress has been reset.", false, "navigation help"))
	for _, text := range []string{"Settings", "Reset lesson progress", "Back", "Lesson progress has been reset."} {
		if !strings.Contains(view, text) {
			t.Errorf("settings view does not contain %q", text)
		}
	}
}

func TestResetConfirmationDefaultsToNo(t *testing.T) {
	view := ansi.Strip(ResetConfirmation(
		[]string{"No, go back", "Yes, reset lesson progress"},
		0,
		false,
		"navigation help",
	))
	for _, text := range []string{"Reset lesson progress?", "cannot be undone", "▶ No, go back", "  Yes, reset lesson progress"} {
		if !strings.Contains(view, text) {
			t.Errorf("reset confirmation does not contain %q", text)
		}
	}
}

func TestCursorStyleMarksCurrentChoiceAndListsBackLast(t *testing.T) {
	view := ansi.Strip(CursorStyle([]string{"Block (blinking)", "Bar (steady)"}, 1, 0, "navigation help"))
	for _, text := range []string{"Cursor style", "▶ Block (blinking)", "  ✓ Bar (steady)"} {
		if !strings.Contains(view, text) {
			t.Errorf("cursor style view does not contain %q", text)
		}
	}
	if strings.Contains(view, "✓ Block (blinking)") {
		t.Error("cursor style view marks a choice that is not current")
	}
	if back, bar := strings.Index(view, "Back"), strings.Index(view, "Bar (steady)"); back < bar {
		t.Error("cursor style view does not list Back after the choices")
	}
}

func TestThemeMarksCurrentChoiceAndListsBackLast(t *testing.T) {
	view := ansi.Strip(Theme([]string{"Dark", "Light", "High contrast"}, 1, 0, "navigation help"))
	for _, text := range []string{"Theme", "▶ Dark", "  ✓ Light", "and the text in the sandbox terminal"} {
		if !strings.Contains(view, text) {
			t.Errorf("theme view does not contain %q", text)
		}
	}
	if strings.Contains(view, "✓ Dark") || strings.Contains(view, "✓ High contrast") {
		t.Error("theme view marks a choice that is not current")
	}
	if back, last := strings.Index(view, "Back"), strings.Index(view, "High contrast"); back < last {
		t.Error("theme view does not list Back after the choices")
	}
}

func TestTerminalProblemViewShowsSummaryRemedyAndSecondaryDetails(t *testing.T) {
	view := ansi.Strip(TerminalProblem{
		Summary: "Docker isn't reachable.",
		Remedy:  "Start Docker.",
		Hint:    "Then press Ctrl+D to go back and try again.",
		Detail:  "Docker said: no daemon",
		LogPath: "/home/learner/shellforge.log",
	}.View())

	order := []string{
		"Oops! There was an error.", "Docker isn't reachable.", "Start Docker.",
		"Then press Ctrl+D to go back and try again.", "Docker said: no daemon", "Full details: /home/learner/shellforge.log",
	}
	last := -1
	for _, text := range order {
		index := strings.Index(view, text)
		if index < 0 || index < last {
			t.Fatalf("view does not show %q after the previous item:\n%s", text, view)
		}
		last = index
	}
}

func TestTerminalProblemViewOmitsEmptySections(t *testing.T) {
	view := ansi.Strip(TerminalProblem{Summary: "Something failed."}.View())

	if want := "Oops! There was an error.\n\nSomething failed."; view != want {
		t.Fatalf("view = %q, want only the title and summary %q", view, want)
	}
}

func TestResetSettingsConfirmationDefaultsToNoAndKeepsProgress(t *testing.T) {
	view := ansi.Strip(ResetSettingsConfirmation(
		[]string{"No, go back", "Yes, reset settings"},
		0,
		"navigation help",
	))
	for _, text := range []string{"Reset settings?", "Lesson progress is not affected", "▶ No, go back", "  Yes, reset settings"} {
		if !strings.Contains(view, text) {
			t.Errorf("settings reset confirmation does not contain %q", text)
		}
	}
}

func TestManageSandboxesStates(t *testing.T) {
	listedAt := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		screen ManageSandboxes
		want   []string
	}{
		{name: "loading", screen: ManageSandboxes{Loading: true, Spinner: "⣾"}, want: []string{"Looking for Shellforge sandboxes... ⣾"}},
		{name: "empty", screen: ManageSandboxes{}, want: []string{"No Shellforge sandboxes found."}},
		{name: "problem", screen: ManageSandboxes{Problem: "Docker is not running"}, want: []string{"Docker is not running"}},
		{
			name: "rows",
			screen: ManageSandboxes{
				ListedAt: listedAt,
				Sandboxes: []container.ManagedSandbox{
					{Name: "shellforge-a", State: "running", CreatedAt: listedAt.Add(-3 * time.Hour)},
					{Name: "shellforge-bb", State: "exited", CreatedAt: listedAt.Add(-10 * time.Second)},
				},
			},
			want: []string{"shellforge-a   running  3h ago", "shellforge-bb  exited   just now"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.screen.Help = "navigation help"
			view := ansi.Strip(test.screen.View())
			for _, want := range append(test.want, config.SettingsItemManageSandboxes, "▶ "+config.BackMenuLabel, "navigation help") {
				if !strings.Contains(view, want) {
					t.Errorf("view does not contain %q:\n%s", want, view)
				}
			}
		})
	}
}

func TestSandboxAgeUsesLargestWholeUnit(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		elapsed time.Duration
		want    string
	}{
		{elapsed: 59 * time.Second, want: "just now"},
		{elapsed: time.Minute, want: "1m ago"},
		{elapsed: 59*time.Minute + 59*time.Second, want: "59m ago"},
		{elapsed: time.Hour, want: "1h ago"},
		{elapsed: 23 * time.Hour, want: "23h ago"},
		{elapsed: 24 * time.Hour, want: "1d ago"},
	}
	for _, test := range tests {
		if got := age(now.Add(-test.elapsed), now); got != test.want {
			t.Errorf("age(%v) = %q, want %q", test.elapsed, got, test.want)
		}
	}
}
