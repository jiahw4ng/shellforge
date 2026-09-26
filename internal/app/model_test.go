package app

import (
	"context"
	"errors"
	"fmt"
	"shellforge/internal/assertion"
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
	for _, text := range append([]string{"Welcome to Shellforge!"}, menuItems...) {
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

func TestMenuNavigationStopsAtBounds(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyUp, ""))
	if model.nav.selection != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.nav.selection)
	}

	for range menuItems {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	if model.nav.selection != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.nav.selection, len(menuItems)-1)
	}
}

func TestEnterShowsSettingsAndBackReturnsToMenu(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != settingsScreen {
		t.Fatalf("screen = %d after enter, want settings screen", model.nav.screen)
	}
	if !strings.Contains(model.View().Content, "Reset lesson progress") {
		t.Fatal("settings view does not contain reset option")
	}

	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.nav.screen)
	}
	if model.nav.selection != 0 {
		t.Fatalf("selected = %d after back, want 0", model.nav.selection)
	}
}

func TestResetLessonProgressRequiresConfirmation(t *testing.T) {
	model := State{nav: navigationState{screen: settingsScreen}}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != resetConfirmationScreen {
		t.Fatalf("screen = %d after choosing reset, want confirmation screen", model.nav.screen)
	}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("default confirmation choice started a reset")
	}
	if result.nav.screen != settingsScreen {
		t.Fatalf("screen = %d after declining reset, want settings screen", result.nav.screen)
	}
}

func TestConfirmedResetClearsPersistedAndInMemoryCompletion(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	model.nav = navigationState{screen: resetConfirmationScreen, selection: 1}
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
	if result.nav.screen != settingsScreen || result.settings.message != "Lesson progress has been reset." {
		t.Fatalf("reset result state = %#v", result.settings)
	}
}

func TestFailedResetPreservesCompletion(t *testing.T) {
	store := &fakeCompletionStore{resetErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	model.nav = navigationState{screen: resetConfirmationScreen, selection: 1}
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
	model.nav = navigationState{screen: resetConfirmationScreen, selection: 1}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("reset without storage returned a command")
	}
	if result.nav.screen != settingsScreen || !result.settings.failed {
		t.Fatalf("reset without storage state = %#v", result.settings)
	}
}

func TestInitLoadsEmbeddedLessons(t *testing.T) {
	model := New()
	model = updateModel(t, model, model.Init()())

	if model.lessons.err != nil {
		t.Fatalf("lesson load error = %v", model.lessons.err)
	}
	if len(model.lessons.available) != 3 {
		t.Fatalf("loaded lessons = %d, want 3", len(model.lessons.available))
	}
	if model.lessons.available[0].ID != "00-introduction" {
		t.Fatalf("first lesson ID = %q, want 00-introduction", model.lessons.available[0].ID)
	}
}

func TestInitLoadsPersistedCompletions(t *testing.T) {
	store := &fakeCompletionStore{lessonIDs: []string{"00-introduction"}}
	model := NewWithCompletionStore(store)
	message := model.Init()()
	batch, ok := message.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init() returned %T, want tea.BatchMsg", message)
	}
	for _, command := range batch {
		model = updateModel(t, model, command())
	}

	if len(model.lessons.available) != 3 {
		t.Fatalf("loaded lessons = %d, want 3", len(model.lessons.available))
	}
	if !model.lessons.completed["00-introduction"] {
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

	if model.nav.screen != lessonsScreen {
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
	model.nav.screen = lessonsScreen
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.nav.screen != lessonScreen {
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
		nav: navigationState{screen: lessonScreen},
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

func TestF12DoesNotStartAssertionsWithoutATerminal(t *testing.T) {
	model := State{nav: navigationState{screen: lessonScreen}, lessons: lessonState{available: []lessons.Lesson{{}}}}
	updated, command := model.Update(keyPress(tea.KeyF12, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("F12 without a terminal returned an assertion command")
	}
	if result.lessons.progress.isChecking {
		t.Fatal("F12 without a terminal started an assertion check")
	}
}

func TestCtrlAltRRestartsLessonAfterTerminalError(t *testing.T) {
	model := New()
	model.nav.screen = lessonScreen
	model.lessons.available = []lessons.Lesson{{ID: "lesson-id"}}
	model.lessons.completed["lesson-id"] = true
	model.lessons.progress = progressState{hasChecked: true, results: []assertion.Result{{Passed: false}}}
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
	if !result.lessons.completed["lesson-id"] {
		t.Fatal("Ctrl+Alt+R cleared persisted lesson completion")
	}
}

func TestCtrlAltRIgnoresResetWhileTerminalStarts(t *testing.T) {
	model := New()
	model.nav.screen = lessonScreen
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
	model.nav.screen = lessonScreen
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
		nav: navigationState{screen: lessonScreen},
		term: terminalState{
			hasRequestedExit: true,
			output:           "old terminal text",
			err:              errors.New("old terminal error"),
		},
	}
	model = updateModel(t, model, termExitedMsg{})
	if model.nav.screen != lessonsScreen {
		t.Fatalf("screen = %d after lesson terminal exit, want lessons screen", model.nav.screen)
	}
	if model.term.output != "" || model.term.err != nil {
		t.Fatalf("terminal state was not cleared: output=%q error=%v", model.term.output, model.term.err)
	}
}

func TestTerminalStartWithoutSessionOrErrorShowsFailure(t *testing.T) {
	model := State{nav: navigationState{screen: lessonScreen}}
	model = updateModel(t, model, termStartedMsg{})
	if model.term.err == nil {
		t.Fatal("terminal error = nil, want invalid-start-result error")
	}
	if !strings.Contains(model.term.err.Error(), "no session and no error") {
		t.Fatalf("terminal error = %q, want invalid-start-result error", model.term.err)
	}
}

func TestUnexpectedLessonTerminalExitStaysOnLesson(t *testing.T) {
	model := State{nav: navigationState{screen: lessonScreen}}
	model = updateModel(t, model, termExitedMsg{})
	if model.nav.screen != lessonScreen {
		t.Fatalf("screen = %d after unexpected terminal exit, want lesson screen", model.nav.screen)
	}
	if model.term.err == nil {
		t.Fatal("terminal error = nil after unexpected terminal exit")
	}

	model = updateModel(t, model, keyPress('d', "", tea.ModCtrl))
	if model.nav.screen != lessonsScreen {
		t.Fatalf("screen = %d after Ctrl+D acknowledgement, want lessons screen", model.nav.screen)
	}
}

func TestLessonBackReturnsToMainMenu(t *testing.T) {
	model := loadedState(t)
	model.nav.screen = lessonsScreen
	model.nav.selection = len(model.lessons.available)
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.nav.screen != menuScreen {
		t.Fatalf("screen = %d after selecting lesson Back, want menu screen", model.nav.screen)
	}
	if model.nav.selection != 0 {
		t.Fatalf("selection = %d after selecting lesson Back, want 0", model.nav.selection)
	}
}

func TestStartLearningRequestsTerminal(t *testing.T) {
	updated, command := New().Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.nav.screen != terminalScreen {
		t.Fatalf("screen = %d after selecting Sandbox, want terminal screen", result.nav.screen)
	}
	if command == nil {
		t.Fatal("selecting Sandbox returned no terminal start command")
	}
}

func TestCtrlCQuitsFromMenuAndSettings(t *testing.T) {
	for _, model := range []State{New(), {nav: navigationState{screen: settingsScreen}}} {
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
	model.nav.selection = len(menuItems) - 1
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

func TestTerminalDimensionsStayInsideFrame(t *testing.T) {
	sandbox := State{nav: navigationState{screen: terminalScreen}, viewport: viewportState{width: 100, height: 40}}
	if width, height := sandbox.terminalDimensions(); width != 96 || height != 38 {
		t.Fatalf("sandbox terminal dimensions = %dx%d, want 96x38", width, height)
	}

	lesson := State{nav: navigationState{screen: lessonScreen}, viewport: viewportState{width: 100, height: 40}}
	if width, height := lesson.terminalDimensions(); width != 47 || height != 38 {
		t.Fatalf("lesson terminal dimensions = %dx%d, want 47x38", width, height)
	}
}

func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := State{nav: navigationState{screen: terminalScreen}, term: terminalState{hasRequestedExit: true}}
	updated, command := model.Update(termExitedMsg{})
	result := updated.(State)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.nav.screen != menuScreen {
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
