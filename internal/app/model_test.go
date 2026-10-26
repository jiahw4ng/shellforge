package app

import (
	"context"
	"errors"
	"fmt"
	"shellforge/internal/assertion"
	"shellforge/internal/config"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestInitialViewShowsWelcomeAndMenu(t *testing.T) {
	view := ansi.Strip(New().View().Content)
	for _, text := range append([]string{"Welcome to Shellforge!"}, config.MenuItems...) {
		if !strings.Contains(view, text) {
			t.Errorf("initial view does not contain %q", text)
		}
	}
}

func TestNewInitializesTerminalSpinner(t *testing.T) {
	model := New()
	if len(model.term.spinner.Spinner.Frames) != len(spinner.Dot.Frames) {
		t.Fatal("New() did not initialize the terminal spinner")
	}
}

func TestF3TogglesHelpOutsideSandbox(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyF3, ""))
	if !model.help.ShowAll {
		t.Fatal("F3 did not expand help")
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "▼ F3 Help") {
		t.Fatalf("expanded help view = %q, want hide disclosure", view)
	}

	model = updateModel(t, model, keyPress(tea.KeyF3, ""))
	if model.help.ShowAll {
		t.Fatal("second F3 did not collapse help")
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "▶ F3 Help") {
		t.Fatalf("collapsed help view = %q, want show disclosure", view)
	}
}

func TestF3DoesNotToggleHelpInSandbox(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model = updateModel(t, model, keyPress(tea.KeyF3, ""))
	if model.help.ShowAll {
		t.Fatal("F3 toggled app help in the sandbox terminal")
	}
}

func TestNavigationStopsAtBounds(t *testing.T) {
	tests := []struct {
		name      string
		model     State
		itemCount int
	}{
		{name: "menu", model: New(), itemCount: len(config.MenuItems)},
		{
			name: "lessons",
			model: State{
				nav:     navigationState{screen: config.LessonsScreen},
				lessons: lessonState{available: []lessons.Lesson{{}, {}}},
			},
			itemCount: 3,
		},
		{
			name:      "settings",
			model:     State{nav: navigationState{screen: config.SettingsScreen}},
			itemCount: len(config.SettingsItems),
		},
		{
			name:      "cursor style",
			model:     State{nav: navigationState{screen: config.CursorStyleScreen}},
			itemCount: len(config.CursorStyles) + 1,
		},
		{
			name:      "reset confirmation",
			model:     State{nav: navigationState{screen: config.ResetConfirmationScreen}},
			itemCount: len(config.ResetConfirmationItems),
		},
		{
			name:      "lessons loading",
			model:     State{nav: navigationState{screen: config.LessonsScreen}},
			itemCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := updateModel(t, test.model, keyPress(tea.KeyUp, ""))
			if model.nav.selection != 0 {
				t.Fatalf("selected = %d after moving up from first item, want 0", model.nav.selection)
			}

			for range test.itemCount + 1 {
				model = updateModel(t, model, keyPress(tea.KeyDown, ""))
			}
			want := max(test.itemCount-1, 0)
			if model.nav.selection != want {
				t.Fatalf("selected = %d after moving down, want %d", model.nav.selection, want)
			}
		})
	}
}

func TestEnterShowsSettingsAndBackReturnsToMenu(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.SettingsScreen {
		t.Fatalf("screen = %d after enter, want settings screen", model.nav.screen)
	}
	if !strings.Contains(model.View().Content, "Reset lesson progress") {
		t.Fatal("settings view does not contain reset option")
	}

	for range len(config.SettingsItems) - 1 {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != config.MenuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.nav.screen)
	}
	if model.nav.selection != 0 {
		t.Fatalf("selected = %d after back, want 0", model.nav.selection)
	}
}

func TestResetLessonProgressRequiresConfirmation(t *testing.T) {
	model := State{nav: navigationState{screen: config.SettingsScreen, selection: settingsResetIndex}}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != config.ResetConfirmationScreen {
		t.Fatalf("screen = %d after choosing reset, want confirmation screen", model.nav.screen)
	}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("default confirmation choice started a reset")
	}
	if result.nav.screen != config.SettingsScreen {
		t.Fatalf("screen = %d after declining reset, want settings screen", result.nav.screen)
	}
}

func TestConfirmedResetClearsPersistedAndInMemoryCompletion(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	model.nav = navigationState{screen: config.ResetConfirmationScreen, selection: 1}
	model.lessons.completed["00-introduction"] = true

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command == nil || !result.settings.isResetting {
		t.Fatal("confirmed reset did not start an asynchronous reset")
	}
	message := command()
	if store.resetCalls != 1 {
		t.Fatalf("reset calls = %d, want 1", store.resetCalls)
	}
	result = updateModel(t, result, message)
	if len(result.lessons.completed) != 0 {
		t.Fatalf("completed lessons after reset = %v, want none", result.lessons.completed)
	}
	if result.nav.screen != config.SettingsScreen || result.settings.message != "Lesson progress has been reset." {
		t.Fatalf("reset result state = %#v", result.settings)
	}
}

func TestFailedResetPreservesCompletion(t *testing.T) {
	store := &fakeCompletionStore{resetErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	model.nav = navigationState{screen: config.ResetConfirmationScreen, selection: 1}
	model.lessons.completed["00-introduction"] = true

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	result = updateModel(t, result, command())
	if !result.lessons.completed["00-introduction"] {
		t.Fatal("failed reset removed in-memory completion")
	}
	if !result.settings.failed || result.settings.message == "" {
		t.Fatalf("failed reset status = %#v", result.settings)
	}
}

func TestResetReportsUnavailableStorage(t *testing.T) {
	model := New()
	model.nav = navigationState{screen: config.ResetConfirmationScreen, selection: 1}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("reset without storage returned a command")
	}
	if result.nav.screen != config.SettingsScreen || !result.settings.failed {
		t.Fatalf("reset without storage state = %#v", result.settings)
	}
}

func TestInitLoadsEmbeddedLessons(t *testing.T) {
	model := New()
	model = updateModel(t, model, model.Init()())

	if model.lessons.err != nil {
		t.Fatalf("lesson load error = %v", model.lessons.err)
	}
	if len(model.lessons.available) != 4 {
		t.Fatalf("loaded lessons = %d, want 4", len(model.lessons.available))
	}
	if model.lessons.available[0].ID != "00-debugging" {
		t.Fatalf("first lesson ID = %q, want 00-debugging", model.lessons.available[0].ID)
	}
}

func TestInitLoadsPersistedCompletions(t *testing.T) {
	store := &fakeCompletionStore{lessonIDs: []string{"00-debugging"}}
	model := NewWithCompletionStore(store)
	message := model.Init()()
	batch, ok := message.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() returned %T, want tea.BatchMsg", message)
	}
	for _, command := range batch {
		model = updateModel(t, model, command())
	}

	if len(model.lessons.available) != 4 {
		t.Fatalf("loaded lessons = %d, want 4", len(model.lessons.available))
	}
	if !model.lessons.completed["00-debugging"] {
		t.Fatal("persisted completion was not loaded")
	}
}

func TestCompletionLoadFailureLeavesAppUsable(t *testing.T) {
	store := &fakeCompletionStore{loadErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	message := loadCompletedLessons(store)()
	model = updateModel(t, model, message)

	if model.lessons.completed == nil {
		t.Fatal("completion load failure removed initialized completion state")
	}
	if !strings.Contains(ansi.Strip(model.View().Content), "Welcome to Shellforge!") {
		t.Fatal("completion load failure made the main menu unusable")
	}
}

func TestChooseLessonShowsLoadedLessons(t *testing.T) {
	model := loadedState(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.LessonsScreen {
		t.Fatalf("screen = %d after choosing lessons, want lessons screen", model.nav.screen)
	}
	if model.nav.selection != 0 {
		t.Fatalf("lesson selection = %d after opening lessons, want 0", model.nav.selection)
	}
	view := model.View().Content
	for _, lesson := range model.lessons.available {
		want := fmt.Sprintf("%d. %s", lesson.Number, lesson.Title)
		if !strings.Contains(view, want) {
			t.Errorf("lesson view does not contain %q", want)
		}
	}
	if !strings.Contains(view, "Back") {
		t.Error("lesson view does not contain Back")
	}
}

func TestLessonStartsSandbox(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = config.LessonsScreen
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.nav.screen != config.LessonTerminalScreen {
		t.Fatalf("screen = %d after selecting a lesson, want lesson screen", result.nav.screen)
	}
	if command == nil {
		t.Fatal("selecting a lesson returned no terminal start command")
	}
	if !result.term.isStarting {
		t.Fatal("terminal startup spinner was not enabled")
	}
}

func TestLessonPageNavigationUsesCtrlPN(t *testing.T) {
	model := State{
		nav: navigationState{screen: config.LessonTerminalScreen},
		lessons: lessonState{available: []lessons.Lesson{{
			Pages: []lessons.Page{{Title: "pwd", Content: "first"}, {Title: "ls", Content: "second"}},
		}}},
	}

	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.lessons.activePage != 1 {
		t.Fatalf("active page = %d after Ctrl+N, want 1", model.lessons.activePage)
	}
	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.lessons.activePage != 1 {
		t.Fatalf("active page = %d beyond final page, want 1", model.lessons.activePage)
	}
	model = updateModel(t, model, keyPress('p', "", tea.ModCtrl))
	if model.lessons.activePage != 0 {
		t.Fatalf("active page = %d after Ctrl+P, want 0", model.lessons.activePage)
	}
}

func TestLessonGuideScrollsWithPageUpAndPageDown(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 100, height: 24}
	model.lessons.available = []lessons.Lesson{{
		Pages: []lessons.Page{{
			Title:   "long guide",
			Content: lessons.Markdown(strings.Repeat("guide line\n\n", 40)),
		}},
	}}
	model.prepareLessonGuide()

	model = updateModel(t, model, keyPress(tea.KeyPgDown, ""))
	if model.lessons.guide.YOffset() != 1 {
		t.Fatalf("guide offset after PgDn = %d, want 1", model.lessons.guide.YOffset())
	}

	model = updateModel(t, model, keyPress(tea.KeyPgUp, ""))
	if model.lessons.guide.YOffset() != 0 {
		t.Fatalf("guide offset after PgUp = %d, want 0", model.lessons.guide.YOffset())
	}
}

func TestChangingLessonPageResetsGuideScroll(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 100, height: 24}
	model.lessons.available = []lessons.Lesson{{Pages: []lessons.Page{
		{Title: "first", Content: lessons.Markdown(strings.Repeat("first page\n\n", 40))},
		{Title: "second", Content: "second page"},
	}}}
	model.prepareLessonGuide()
	model.lessons.guide.PageDown()
	if model.lessons.guide.YOffset() == 0 {
		t.Fatal("test setup did not scroll the guide")
	}

	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.lessons.guide.YOffset() != 0 {
		t.Fatalf("guide offset after changing page = %d, want 0", model.lessons.guide.YOffset())
	}
}

func TestChangingLessonPageCollapsesHelp(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{Pages: []lessons.Page{
		{Title: "first"},
		{Title: "second"},
	}}}
	model.help.ShowAll = true

	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.help.ShowAll {
		t.Fatal("changing to the next page left help expanded")
	}

	model.help.ShowAll = true
	model = updateModel(t, model, keyPress('p', "", tea.ModCtrl))
	if model.help.ShowAll {
		t.Fatal("changing to the previous page left help expanded")
	}

	model.help.ShowAll = true
	model = updateModel(t, model, keyPress('p', "", tea.ModCtrl))
	if !model.help.ShowAll {
		t.Fatal("pressing previous on the first page collapsed help without changing pages")
	}
}

func TestF10RevealsHintsOneAtATimeAndExpandsThem(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 100, height: 24}
	model.lessons.available = []lessons.Lesson{{
		Pages: []lessons.Page{{
			Title:   "long guide",
			Content: lessons.Markdown(strings.Repeat("guide line\n\n", 40)),
		}},
		Hints: []string{"First hint.", "Second hint."},
	}}

	model = updateModel(t, model, keyPress(tea.KeyF10, ""))
	if model.lessons.revealedHints != 1 {
		t.Fatalf("revealed hints after first F10 = %d, want 1", model.lessons.revealedHints)
	}
	if !model.lessons.hintsExpanded {
		t.Fatal("F10 did not expand revealed hints")
	}

	model = updateModel(t, model, keyPress(tea.KeyF10, ""))
	model = updateModel(t, model, keyPress(tea.KeyF10, ""))
	if model.lessons.revealedHints != 2 {
		t.Fatalf("revealed hints after repeated F10 = %d, want 2", model.lessons.revealedHints)
	}
}

func TestF1TogglesHintsBeforeAndAfterReveal(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{Pages: []lessons.Page{{}}, Hints: []string{"A hint."}}}

	model = updateModel(t, model, keyPress(tea.KeyF1, ""))
	if !model.lessons.hintsExpanded {
		t.Fatal("F1 did not expand hints before one was revealed")
	}
	model = updateModel(t, model, keyPress(tea.KeyF1, ""))
	if model.lessons.hintsExpanded {
		t.Fatal("second F1 did not collapse empty hints")
	}

	model.lessons.revealedHints = 1
	model = updateModel(t, model, keyPress(tea.KeyF1, ""))
	if !model.lessons.hintsExpanded || model.lessons.revealedHints != 1 {
		t.Fatalf("hint state after F1 = expanded %t, revealed %d", model.lessons.hintsExpanded, model.lessons.revealedHints)
	}
	model = updateModel(t, model, keyPress(tea.KeyF1, ""))
	if model.lessons.hintsExpanded {
		t.Fatal("second F1 did not collapse hints")
	}
}

func TestRevealedHintsResetOnlyWhenLessonIsReopened(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{
		Pages: []lessons.Page{{Title: "first"}, {Title: "second"}},
		Hints: []string{"A hint."},
	}}
	model.lessons.revealedHints = 1
	model.lessons.hintsExpanded = true

	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.lessons.revealedHints != 1 {
		t.Fatal("changing lesson pages reset revealed hints")
	}
	if !model.lessons.hintsExpanded {
		t.Fatal("changing lesson pages collapsed hints")
	}

	model.nav.screen = config.LessonsScreen
	model.nav.selection = 0
	model.handleEnter()
	if model.lessons.revealedHints != 0 {
		t.Fatalf("revealed hints after reopening lesson = %d, want 0", model.lessons.revealedHints)
	}
	if model.lessons.hintsExpanded {
		t.Fatal("reopening lesson left hints expanded")
	}
}

func TestF10WithoutHintsIsConsumedAsANoOp(t *testing.T) {
	model := State{
		nav:     navigationState{screen: config.LessonTerminalScreen},
		lessons: lessonState{available: []lessons.Lesson{{Pages: []lessons.Page{{}}}}},
	}

	updated, command := model.Update(keyPress(tea.KeyF10, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("F10 without hints returned a command")
	}
	if result.lessons.revealedHints != 0 {
		t.Fatalf("revealed hints without available hints = %d, want 0", result.lessons.revealedHints)
	}
}

func TestF12DoesNotStartAssertionsWithoutATerminal(t *testing.T) {
	model := State{nav: navigationState{screen: config.LessonTerminalScreen}, lessons: lessonState{available: []lessons.Lesson{{}}}}
	updated, command := model.Update(keyPress(tea.KeyF12, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("F12 without a terminal returned an assertion command")
	}
	if result.lessons.progress.isChecking {
		t.Fatal("F12 without a terminal started an assertion check")
	}
}

func TestAssertionResultsExpandAndF2TogglesThem(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{Pages: []lessons.Page{{}}}}

	model = updateModel(t, model, keyPress(tea.KeyF2, ""))
	if !model.lessons.progress.assertionsExpanded {
		t.Fatal("F2 did not expand assertions before a check completed")
	}
	model = updateModel(t, model, keyPress(tea.KeyF2, ""))
	if model.lessons.progress.assertionsExpanded {
		t.Fatal("second F2 did not collapse empty assertions")
	}

	model.lessons.progress.isChecking = true
	model = updateModel(t, model, assertionsCheckedMsg{
		results: []assertion.Result{{Passed: false, Message: "Missing file."}},
	})
	if !model.lessons.progress.hasChecked || !model.lessons.progress.assertionsExpanded {
		t.Fatalf("completed assertion state = %#v, want checked and expanded", model.lessons.progress)
	}

	model = updateModel(t, model, keyPress(tea.KeyF2, ""))
	if model.lessons.progress.assertionsExpanded {
		t.Fatal("F2 did not collapse completed assertions")
	}
	model = updateModel(t, model, keyPress(tea.KeyF2, ""))
	if !model.lessons.progress.assertionsExpanded {
		t.Fatal("second F2 did not expand completed assertions")
	}

	model.lessons.progress.isChecking = true
	model = updateModel(t, model, keyPress(tea.KeyF2, ""))
	if model.lessons.progress.assertionsExpanded {
		t.Fatal("F2 did not collapse assertions while a check was running")
	}
}

func TestCtrlAltRRestartsLessonAfterTerminalError(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{ID: "lesson-id", Hints: []string{"A hint."}}}
	model.lessons.completed["lesson-id"] = true
	model.lessons.revealedHints = 1
	model.lessons.hintsExpanded = true
	model.lessons.progress = progressState{
		hasChecked:         true,
		assertionsExpanded: true,
		results:            []assertion.Result{{Passed: false}},
	}
	model.term.err = errors.New("Docker is unavailable")
	model.term.output = "old terminal error"

	updated, command := model.Update(keyPress('r', "", tea.ModCtrl, tea.ModAlt))
	result := updated.(State)

	if command == nil {
		t.Fatal("Ctrl+Alt+R returned no terminal start command")
	}
	if !result.term.isStarting || result.term.err != nil || result.term.output != "" {
		t.Fatalf("terminal state after Ctrl+Alt+R = %#v", result.term)
	}
	if result.term.gen != 1 {
		t.Fatalf("terminal generation = %d, want 1", result.term.gen)
	}
	if result.lessons.progress.hasChecked || len(result.lessons.progress.results) != 0 {
		t.Fatalf("progress after Ctrl+Alt+R = %#v, want reset progress", result.lessons.progress)
	}
	if result.lessons.progress.assertionsExpanded {
		t.Fatal("Ctrl+Alt+R left assertion results expanded")
	}
	if !result.lessons.completed["lesson-id"] {
		t.Fatal("Ctrl+Alt+R cleared persisted lesson completion")
	}
	if result.lessons.revealedHints != 1 {
		t.Fatal("Ctrl+Alt+R cleared revealed hints")
	}
	if !result.lessons.hintsExpanded {
		t.Fatal("Ctrl+Alt+R collapsed revealed hints")
	}
}

func TestCtrlAltRIgnoresResetWhileTerminalStarts(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.lessons.available = []lessons.Lesson{{ID: "lesson-id"}}
	model.term.isStarting = true
	model.term.gen = 7

	updated, command := model.Update(keyPress('r', "", tea.ModCtrl, tea.ModAlt))
	result := updated.(State)

	if command != nil {
		t.Fatal("Ctrl+Alt+R during startup returned another terminal start command")
	}
	if result.term.gen != 7 {
		t.Fatalf("terminal generation = %d, want 7", result.term.gen)
	}
}

func TestStaleTerminalMessagesDoNotChangeNewAttempt(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.term.gen = 2
	model.term.isStarting = true
	model.lessons.progress.isChecking = true

	model = updateModel(t, model, termExitedMsg{gen: 1})
	if !model.term.isStarting || model.term.err != nil {
		t.Fatalf("stale terminal exit changed terminal state: %#v", model.term)
	}

	model = updateModel(t, model, assertionsCheckedMsg{
		lessonID: "old-lesson",
		results:  []assertion.Result{{Passed: true}},
		gen:      1,
	})
	if !model.lessons.progress.isChecking || len(model.lessons.progress.results) != 0 {
		t.Fatalf("stale assertion result changed progress: %#v", model.lessons.progress)
	}
}

func TestCompletionsLoadedMergesWithInSessionCompletion(t *testing.T) {
	model := New()
	model.lessons.completed["in-session"] = true
	model = updateModel(t, model, completionsLoadedMsg{lessonIDs: []string{"persisted"}})

	for _, lessonID := range []string{"in-session", "persisted"} {
		if !model.lessons.completed[lessonID] {
			t.Fatalf("lesson %q was not marked complete", lessonID)
		}
	}
}

func TestPassingAssertionsPersistLessonCompletion(t *testing.T) {
	for _, test := range []struct {
		name    string
		results []assertion.Result
	}{
		{name: "all assertions pass", results: []assertion.Result{{Passed: true}, {Passed: true}}},
		{name: "lesson has no assertions", results: nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeCompletionStore{}
			model := NewWithCompletionStore(store)
			updated, command := model.Update(assertionsCheckedMsg{lessonID: "lesson-id", results: test.results})
			result := updated.(State)

			if !result.lessons.completed["lesson-id"] {
				t.Fatal("passing assertion result did not mark lesson complete")
			}
			if command == nil {
				t.Fatal("passing assertion result returned no persistence command")
			}
			message := command()
			saved, ok := message.(completionSavedMsg)
			if !ok || saved.err != nil {
				t.Fatalf("completion command returned %#v", message)
			}
			if len(store.marked) != 1 || store.marked[0] != "lesson-id" {
				t.Fatalf("persisted lesson IDs = %v, want [lesson-id]", store.marked)
			}
		})
	}
}

func TestFailedAssertionsDoNotCompleteLesson(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	updated, command := model.Update(assertionsCheckedMsg{
		lessonID: "lesson-id",
		results:  []assertion.Result{{Passed: true}, {Passed: false}},
	})
	result := updated.(State)

	if result.lessons.completed["lesson-id"] {
		t.Fatal("failed assertion marked lesson complete")
	}
	if command != nil || len(store.marked) != 0 {
		t.Fatal("failed assertion attempted to persist completion")
	}
}

func TestCompletedLessonDoesNotWriteAgain(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	model.lessons.completed["lesson-id"] = true
	_, command := model.Update(assertionsCheckedMsg{lessonID: "lesson-id"})

	if command != nil || len(store.marked) != 0 {
		t.Fatal("completed lesson attempted a duplicate persistence write")
	}
}

func TestCompletionWriteFailureKeepsInMemoryCompletion(t *testing.T) {
	store := &fakeCompletionStore{markErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	updated, command := model.Update(assertionsCheckedMsg{lessonID: "lesson-id"})
	result := updated.(State)

	if !result.lessons.completed["lesson-id"] {
		t.Fatal("write failure removed in-memory completion")
	}
	saved := command().(completionSavedMsg)
	if saved.err == nil {
		t.Fatal("completion command did not report write failure")
	}
	updated, followUp := result.Update(saved)
	if followUp != nil || !updated.(State).lessons.completed["lesson-id"] {
		t.Fatal("handling write failure changed completion state")
	}
}

func TestLessonExitReturnsToLessons(t *testing.T) {
	model := State{
		nav: navigationState{screen: config.LessonTerminalScreen},
		term: terminalState{
			hasRequestedExit: true,
			output:           "old terminal text",
			err:              errors.New("old terminal error"),
		},
	}
	model = updateModel(t, model, termExitedMsg{})
	if model.nav.screen != config.LessonsScreen {
		t.Fatalf("screen = %d after lesson terminal exit, want lessons screen", model.nav.screen)
	}
	if model.term.output != "" || model.term.err != nil {
		t.Fatalf("terminal state was not cleared: output=%q error=%v", model.term.output, model.term.err)
	}
}

func TestTerminalStartWithoutSessionOrErrorShowsFailure(t *testing.T) {
	model := State{nav: navigationState{screen: config.LessonTerminalScreen}}
	model = updateModel(t, model, termStartedMsg{})
	if model.term.err == nil {
		t.Fatal("terminal error = nil, want invalid-start-result error")
	}
	if !strings.Contains(model.term.err.Error(), "no session and no error") {
		t.Fatalf("terminal error = %q, want invalid-start-result error", model.term.err)
	}
}

func TestLessonTerminalFailureAcknowledgementReturnsToLessons(t *testing.T) {
	model := State{
		nav: navigationState{screen: config.LessonTerminalScreen},
		lessons: lessonState{
			available: []lessons.Lesson{{Pages: []lessons.Page{{}}}},
		},
		term: terminalState{err: errors.New("the terminal closed unexpectedly")},
	}
	model = updateModel(t, model, keyPress('d', "", tea.ModCtrl))
	if model.nav.screen != config.LessonsScreen {
		t.Fatalf("screen = %d after Ctrl+D acknowledgement, want lessons screen", model.nav.screen)
	}
}

func TestLessonBackReturnsToMainMenu(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = config.LessonsScreen
	model.nav.selection = len(model.lessons.available)
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != config.MenuScreen {
		t.Fatalf("screen = %d after selecting lesson Back, want menu screen", model.nav.screen)
	}
	if model.nav.selection != 0 {
		t.Fatalf("selection = %d after selecting lesson Back, want 0", model.nav.selection)
	}
}

func TestStartLearningRequestsTerminal(t *testing.T) {
	updated, command := New().Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.nav.screen != config.SandboxTerminalScreen {
		t.Fatalf("screen = %d after selecting Sandbox, want terminal screen", result.nav.screen)
	}
	if command == nil {
		t.Fatal("selecting Sandbox returned no terminal start command")
	}
}

func TestCtrlCQuitsFromMenuAndSettings(t *testing.T) {
	for _, model := range []State{New(), {nav: navigationState{screen: config.SettingsScreen}}} {
		_, command := model.Update(keyPress('c', "", tea.ModCtrl))
		if command == nil {
			t.Fatal("Ctrl+C returned no quit command")
		}
		if _, ok := command().(tea.QuitMsg); !ok {
			t.Fatalf("Ctrl+C command returned %T, want tea.QuitMsg", command())
		}
	}
}

func TestExitMenuItemQuits(t *testing.T) {
	model := New()
	model.nav.selection = len(config.MenuItems) - 1
	_, command := model.Update(keyPress(tea.KeyEnter, ""))
	if command == nil {
		t.Fatal("selecting Exit returned no quit command")
	}
	if _, ok := command().(tea.QuitMsg); !ok {
		t.Fatalf("Exit command returned %T, want tea.QuitMsg", command())
	}
}

func TestWindowResizeSetsDimensions(t *testing.T) {
	model := updateModel(t, New(), tea.WindowSizeMsg{Width: 100, Height: 40})
	if model.viewport.width != 100 || model.viewport.height != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.viewport.width, model.viewport.height)
	}
}

func TestApplicationFrameDrawsWhiteBorder(t *testing.T) {
	view := ui.WithAppFrame("Shellforge", 30, 8)
	for _, border := range []string{"┌", "┐", "└", "┘"} {
		if !strings.Contains(view, border) {
			t.Errorf("application frame does not contain %q", border)
		}
	}
}

func TestPositionedTerminalCursorUsesSandboxLayout(t *testing.T) {
	model := State{
		nav:      navigationState{screen: config.SandboxTerminalScreen},
		viewport: viewportState{width: 100, height: 40},
	}
	cursor := &tea.Cursor{Position: tea.Position{X: 3, Y: 4}, Shape: tea.CursorUnderline}

	positioned := model.positionedTerminalCursor(cursor)

	if positioned == nil {
		t.Fatal("positionedTerminalCursor() = nil, want visible cursor")
	}
	if positioned.X != 5 || positioned.Y != 6 {
		t.Errorf("positionedTerminalCursor() position = (%d,%d), want (5,6)", positioned.X, positioned.Y)
	}
}

func TestPositionedTerminalCursorUsesLessonLayout(t *testing.T) {
	model := State{
		nav:      navigationState{screen: config.LessonTerminalScreen},
		viewport: viewportState{width: 100, height: 40},
	}
	cursor := &tea.Cursor{Position: tea.Position{X: 3, Y: 4}, Shape: tea.CursorUnderline}

	positioned := model.positionedTerminalCursor(cursor)

	if positioned == nil {
		t.Fatal("positionedTerminalCursor() = nil, want visible cursor")
	}
	if positioned.X != 54 || positioned.Y != 6 {
		t.Errorf("positionedTerminalCursor() position = (%d,%d), want (54,6)", positioned.X, positioned.Y)
	}
}

func TestPositionedTerminalCursorHidesOutsideReadyTerminal(t *testing.T) {
	cursor := &tea.Cursor{}
	tests := []struct {
		name  string
		model State
	}{
		{
			name:  "non-terminal screen",
			model: State{nav: navigationState{screen: config.MenuScreen}, viewport: viewportState{width: 100, height: 40}},
		},
		{
			name:  "missing viewport dimensions",
			model: State{nav: navigationState{screen: config.SandboxTerminalScreen}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if positioned := test.model.positionedTerminalCursor(cursor); positioned != nil {
				t.Errorf("positionedTerminalCursor() = %#v, want nil", positioned)
			}
		})
	}
}

func TestLessonApplicationFrameJoinsPaneDivider(t *testing.T) {
	model := New()
	model.nav.screen = config.LessonTerminalScreen
	model.viewport = viewportState{width: 30, height: 10}
	model.lessons.available = []lessons.Lesson{{
		Number: 1,
		Title:  "Getting around",
		Pages:  []lessons.Page{{Title: "pwd", Content: "Run `pwd`."}},
	}}

	lines := strings.Split(ansi.Strip(model.View().Content), "\n")
	if !strings.Contains(lines[0], "┬") {
		t.Errorf("lesson frame top border does not join pane divider: %q", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "┴") {
		t.Errorf("lesson frame bottom border does not join pane divider: %q", lines[len(lines)-1])
	}
}

func TestTerminalDimensionsStayInsideFrame(t *testing.T) {
	sandbox := State{nav: navigationState{screen: config.SandboxTerminalScreen}, viewport: viewportState{width: 100, height: 40}}
	if width, height := sandbox.terminalDimensions(); width != 96 || height != 37 {
		t.Fatalf("sandbox terminal dimensions = %dx%d, want 96x37", width, height)
	}

	lesson := State{nav: navigationState{screen: config.LessonTerminalScreen}, viewport: viewportState{width: 100, height: 40}}
	if width, height := lesson.terminalDimensions(); width != 47 || height != 37 {
		t.Fatalf("lesson terminal dimensions = %dx%d, want 47x37", width, height)
	}
}

func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := State{nav: navigationState{screen: config.SandboxTerminalScreen}, term: terminalState{hasRequestedExit: true}}
	updated, command := model.Update(termExitedMsg{})
	result := updated.(State)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.nav.screen != config.MenuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.nav.screen)
	}
}

func keyPress(code rune, text string, modifiers ...tea.KeyMod) tea.KeyPressMsg {
	var mod tea.KeyMod
	for _, modifier := range modifiers {
		mod |= modifier
	}
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func updateModel(t *testing.T, model State, message tea.Msg) State {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(State)
	if !ok {
		t.Fatalf("updated model type = %T, want app.State", updated)
	}
	return result
}

func loadedState(t *testing.T) State {
	t.Helper()
	loaded, err := lessons.Load()
	if err != nil {
		t.Fatalf("load embedded lessons: %v", err)
	}
	return State{lessons: lessonState{available: loaded}}
}

type fakeCompletionStore struct {
	lessonIDs  []string
	loadErr    error
	markErr    error
	resetErr   error
	marked     []string
	resetCalls int
}

func (s *fakeCompletionStore) CompletedLessonIDs(context.Context) ([]string, error) {
	return s.lessonIDs, s.loadErr
}

func (s *fakeCompletionStore) MarkCompleted(_ context.Context, lessonID string) error {
	s.marked = append(s.marked, lessonID)
	return s.markErr
}

func (s *fakeCompletionStore) ResetLessonCompletions(context.Context) error {
	s.resetCalls++
	return s.resetErr
}

func TestPositionedTerminalCursorDefaultsToBlinkingBlock(t *testing.T) {
	model := New()
	model.nav.screen = config.SandboxTerminalScreen
	model.viewport = viewportState{width: 100, height: 40}

	positioned := model.positionedTerminalCursor(&tea.Cursor{Shape: tea.CursorBar})
	if positioned == nil {
		t.Fatal("positionedTerminalCursor() = nil, want visible cursor")
	}
	if positioned.Shape != tea.CursorBlock || !positioned.Blink {
		t.Fatalf("default cursor = shape %d blink %v, want blinking block", positioned.Shape, positioned.Blink)
	}
}

func TestChoosingCursorStyleAppliesAndPersistsIt(t *testing.T) {
	for index, style := range config.CursorStyles {
		t.Run(style.ID, func(t *testing.T) {
			store := &fakeSettingsStore{}
			model := NewWithStores(nil, store)
			model.nav = navigationState{screen: config.SettingsScreen, selection: settingsCursorStyleIndex}
			model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
			if model.nav.screen != config.CursorStyleScreen {
				t.Fatalf("screen = %d after choosing cursor style, want cursor style screen", model.nav.screen)
			}
			model.nav.selection = index

			updated, command := model.Update(keyPress(tea.KeyEnter, ""))
			result := updated.(State)
			if command == nil {
				t.Fatal("choosing a cursor style did not schedule persistence")
			}
			result = updateModel(t, result, command())
			if len(store.saved) != 1 || store.saved[0] != style.ID {
				t.Fatalf("saved cursor styles = %v, want [%s]", store.saved, style.ID)
			}
			if result.nav.screen != config.SettingsScreen || result.settings.failed {
				t.Fatalf("state after choosing = screen %d, settings %#v", result.nav.screen, result.settings)
			}
			if view := ansi.Strip(result.View().Content); !strings.Contains(view, "Cursor style: "+style.Label) {
				t.Errorf("settings view does not show %q", style.Label)
			}

			result.nav.screen = config.SandboxTerminalScreen
			result.viewport = viewportState{width: 100, height: 40}
			positioned := result.positionedTerminalCursor(&tea.Cursor{})
			if positioned == nil || positioned.Shape != style.Shape || positioned.Blink != style.Blink {
				t.Fatalf("cursor = %#v, want shape %d blink %v", positioned, style.Shape, style.Blink)
			}
		})
	}
}

func TestCursorStyleBackKeepsCurrentStyle(t *testing.T) {
	model := NewWithStores(nil, &fakeSettingsStore{})
	model.preferences.cursorStyleID = "bar-steady"
	model.nav = navigationState{screen: config.CursorStyleScreen, selection: len(config.CursorStyles)}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("Back from cursor style scheduled work")
	}
	if result.nav.screen != config.SettingsScreen || result.preferences.cursorStyle().ID != "bar-steady" {
		t.Fatalf("after Back: screen %d, cursor style %q", result.nav.screen, result.preferences.cursorStyleID)
	}
}

func TestCursorStyleScreenStartsOnCurrentStyle(t *testing.T) {
	model := New()
	model.preferences.cursorStyleID = "underline-steady"
	model.nav = navigationState{screen: config.SettingsScreen, selection: settingsCursorStyleIndex}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if got := config.CursorStyles[model.nav.selection].ID; got != "underline-steady" {
		t.Fatalf("selected cursor style = %q, want underline-steady", got)
	}
}

func TestInitLoadsPersistedCursorStyle(t *testing.T) {
	store := &fakeSettingsStore{styleID: "underline-blinking", found: true}
	model := NewWithStores(nil, store)
	batch, ok := model.Init()().(tea.BatchMsg)
	if !ok {
		t.Fatal("Init() did not return a batch")
	}
	for _, command := range batch {
		model = updateModel(t, model, command())
	}

	if got := model.preferences.cursorStyle().ID; got != "underline-blinking" {
		t.Fatalf("cursor style = %q, want underline-blinking", got)
	}
}

func TestCursorStyleLoadKeepsDefaultForUnusableValues(t *testing.T) {
	tests := []struct {
		name  string
		store *fakeSettingsStore
	}{
		{name: "unknown", store: &fakeSettingsStore{styleID: "triangle", found: true}},
		{name: "missing", store: &fakeSettingsStore{}},
		{name: "error", store: &fakeSettingsStore{loadErr: errors.New("database unavailable")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := NewWithStores(nil, test.store)
			model = updateModel(t, model, loadCursorStyle(test.store)())
			if got := model.preferences.cursorStyle(); got != config.DefaultCursorStyle {
				t.Fatalf("cursor style = %q, want default", got.ID)
			}
		})
	}
}

func TestLateCursorStyleLoadDoesNotReplaceChoice(t *testing.T) {
	store := &fakeSettingsStore{styleID: "bar-blinking", found: true}
	model := NewWithStores(nil, store)
	model.nav = navigationState{screen: config.CursorStyleScreen, selection: 1}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	model = updateModel(t, model, loadCursorStyle(store)())

	if got := model.preferences.cursorStyle().ID; got != config.CursorStyles[1].ID {
		t.Fatalf("cursor style = %q after late load, want %q", got, config.CursorStyles[1].ID)
	}
}

func TestFailedCursorStyleSaveKeepsChoiceAndReportsFailure(t *testing.T) {
	store := &fakeSettingsStore{saveErr: errors.New("database unavailable")}
	model := NewWithStores(nil, store)
	model.nav = navigationState{screen: config.CursorStyleScreen, selection: 2}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updateModel(t, updated.(State), command())
	if !result.settings.failed || !strings.Contains(result.settings.message, "could not be saved") {
		t.Fatalf("settings after failed save = %#v", result.settings)
	}
	if got := result.preferences.cursorStyle().ID; got != config.CursorStyles[2].ID {
		t.Fatalf("cursor style = %q after failed save, want %q", got, config.CursorStyles[2].ID)
	}
}

func TestCursorStyleWithoutStorageAppliesForSession(t *testing.T) {
	model := New()
	model.nav = navigationState{screen: config.CursorStyleScreen, selection: 3}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("choosing a cursor style without storage returned a command")
	}
	if got := result.preferences.cursorStyle().ID; got != config.CursorStyles[3].ID {
		t.Fatalf("cursor style = %q, want %q", got, config.CursorStyles[3].ID)
	}
	if !result.settings.failed || !strings.Contains(result.settings.message, "storage is unavailable") {
		t.Fatalf("settings without storage = %#v", result.settings)
	}
}

func TestReopeningSettingsKeepsCursorStyle(t *testing.T) {
	model := New()
	model.preferences.cursorStyleID = "bar-steady"
	model.nav = navigationState{screen: config.MenuScreen, selection: 2}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.SettingsScreen {
		t.Fatalf("screen = %d, want settings screen", model.nav.screen)
	}
	if got := model.preferences.cursorStyle().ID; got != "bar-steady" {
		t.Fatalf("cursor style = %q after reopening Settings, want bar-steady", got)
	}
}

type fakeSettingsStore struct {
	styleID string
	found   bool
	loadErr error
	saveErr error
	saved   []string
}

func (s *fakeSettingsStore) CursorStyle(context.Context) (string, bool, error) {
	return s.styleID, s.found, s.loadErr
}

func (s *fakeSettingsStore) SaveCursorStyle(_ context.Context, styleID string) error {
	s.saved = append(s.saved, styleID)
	return s.saveErr
}
