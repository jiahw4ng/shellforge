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
)

func TestInitialViewShowsWelcomeAndMenu(t *testing.T) {
	view := New().View().Content
	for _, text := range append([]string{"Welcome to Shellforge!"}, menuItems...) {
		if !strings.Contains(view, text) {
			t.Errorf("initial view does not contain %q", text)
		}
	}
}

func TestNewInitializesTerminalSpinner(t *testing.T) {
	model := New()
	if len(model.Term.Spinner.Spinner.Frames) != len(spinner.Dot.Frames) {
		t.Fatal("New() did not initialize the terminal spinner")
	}
}

func TestMenuNavigationStopsAtBounds(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyUp, ""))
	if model.Nav.Selection != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.Nav.Selection)
	}

	for range menuItems {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	if model.Nav.Selection != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.Nav.Selection, len(menuItems)-1)
	}
}

func TestEnterShowsSettingsAndBackReturnsToMenu(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.Nav.Screen != settingsScreen {
		t.Fatalf("screen = %d after enter, want settings screen", model.Nav.Screen)
	}
	if !strings.Contains(model.View().Content, "Reset lesson progress") {
		t.Fatal("settings view does not contain reset option")
	}

	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.Nav.Screen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.Nav.Screen)
	}
	if model.Nav.Selection != 0 {
		t.Fatalf("selected = %d after back, want 0", model.Nav.Selection)
	}
}

func TestResetLessonProgressRequiresConfirmation(t *testing.T) {
	model := State{Nav: navigationState{Screen: settingsScreen}}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.Nav.Screen != resetConfirmationScreen {
		t.Fatalf("screen = %d after choosing reset, want confirmation screen", model.Nav.Screen)
	}

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command != nil {
		t.Fatal("default confirmation choice started a reset")
	}
	if result.Nav.Screen != settingsScreen {
		t.Fatalf("screen = %d after declining reset, want settings screen", result.Nav.Screen)
	}
}

func TestConfirmedResetClearsPersistedAndInMemoryCompletion(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	model.Nav = navigationState{Screen: resetConfirmationScreen, Selection: 1}
	model.Lessons.Completed["00-introduction"] = true

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	if command == nil || !result.Settings.isResetting {
		t.Fatal("confirmed reset did not start an asynchronous reset")
	}
	message := command()
	if store.resetCalls != 1 {
		t.Fatalf("reset calls = %d, want 1", store.resetCalls)
	}
	result = updateModel(t, result, message)
	if len(result.Lessons.Completed) != 0 {
		t.Fatalf("completed lessons after reset = %v, want none", result.Lessons.Completed)
	}
	if result.Nav.Screen != settingsScreen || result.Settings.Message != "Lesson progress has been reset." {
		t.Fatalf("reset result state = %#v", result.Settings)
	}
}

func TestFailedResetPreservesCompletion(t *testing.T) {
	store := &fakeCompletionStore{resetErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	model.Nav = navigationState{Screen: resetConfirmationScreen, Selection: 1}
	model.Lessons.Completed["00-introduction"] = true

	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)
	result = updateModel(t, result, command())
	if !result.Lessons.Completed["00-introduction"] {
		t.Fatal("failed reset removed in-memory completion")
	}
	if !result.Settings.Failed || result.Settings.Message == "" {
		t.Fatalf("failed reset status = %#v", result.Settings)
	}
}

func TestResetReportsUnavailableStorage(t *testing.T) {
	model := New()
	model.Nav = navigationState{Screen: resetConfirmationScreen, Selection: 1}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("reset without storage returned a command")
	}
	if result.Nav.Screen != settingsScreen || !result.Settings.Failed {
		t.Fatalf("reset without storage state = %#v", result.Settings)
	}
}

func TestInitLoadsEmbeddedLessons(t *testing.T) {
	model := New()
	model = updateModel(t, model, model.Init()())

	if model.Lessons.Error != nil {
		t.Fatalf("lesson load error = %v", model.Lessons.Error)
	}
	if len(model.Lessons.Available) != 3 {
		t.Fatalf("loaded lessons = %d, want 3", len(model.Lessons.Available))
	}
	if model.Lessons.Available[0].ID != "00-introduction" {
		t.Fatalf("first lesson ID = %q, want 00-introduction", model.Lessons.Available[0].ID)
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

	if len(model.Lessons.Available) != 3 {
		t.Fatalf("loaded lessons = %d, want 3", len(model.Lessons.Available))
	}
	if !model.Lessons.Completed["00-introduction"] {
		t.Fatal("persisted completion was not loaded")
	}
}

func TestCompletionLoadFailureLeavesAppUsable(t *testing.T) {
	store := &fakeCompletionStore{loadErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	message := loadCompletedLessons(store)()
	model = updateModel(t, model, message)

	if model.Lessons.Completed == nil {
		t.Fatal("completion load failure removed initialized completion state")
	}
	if !strings.Contains(model.View().Content, "Welcome to Shellforge!") {
		t.Fatal("completion load failure made the main menu unusable")
	}
}

func TestChooseLessonShowsLoadedLessons(t *testing.T) {
	model := loadedState(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.Nav.Screen != lessonsScreen {
		t.Fatalf("screen = %d after choosing lessons, want lessons screen", model.Nav.Screen)
	}
	if model.Nav.Selection != 0 {
		t.Fatalf("lesson selection = %d after opening lessons, want 0", model.Nav.Selection)
	}
	view := model.View().Content
	for _, lesson := range model.Lessons.Available {
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
	model.Nav.Screen = lessonsScreen
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.Nav.Screen != lessonScreen {
		t.Fatalf("screen = %d after selecting a lesson, want lesson screen", result.Nav.Screen)
	}
	if command == nil {
		t.Fatal("selecting a lesson returned no terminal start command")
	}
	if !result.Term.isStarting {
		t.Fatal("terminal startup spinner was not enabled")
	}
}

func TestLessonPageNavigationUsesCtrlPN(t *testing.T) {
	model := State{
		Nav: navigationState{Screen: lessonScreen},
		Lessons: lessonState{Available: []lessons.Lesson{{
			Pages: []lessons.Page{{Title: "pwd", Content: "first"}, {Title: "ls", Content: "second"}},
		}}},
	}

	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.Lessons.ActivePage != 1 {
		t.Fatalf("active page = %d after Ctrl+N, want 1", model.Lessons.ActivePage)
	}
	model = updateModel(t, model, keyPress('n', "", tea.ModCtrl))
	if model.Lessons.ActivePage != 1 {
		t.Fatalf("active page = %d beyond final page, want 1", model.Lessons.ActivePage)
	}
	model = updateModel(t, model, keyPress('p', "", tea.ModCtrl))
	if model.Lessons.ActivePage != 0 {
		t.Fatalf("active page = %d after Ctrl+P, want 0", model.Lessons.ActivePage)
	}
}

func TestF12DoesNotStartAssertionsWithoutATerminal(t *testing.T) {
	model := State{Nav: navigationState{Screen: lessonScreen}, Lessons: lessonState{Available: []lessons.Lesson{{}}}}
	updated, command := model.Update(keyPress(tea.KeyF12, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("F12 without a terminal returned an assertion command")
	}
	if result.Lessons.Progress.isChecking {
		t.Fatal("F12 without a terminal started an assertion check")
	}
}

func TestCompletionsLoadedMergesWithInSessionCompletion(t *testing.T) {
	model := New()
	model.Lessons.Completed["in-session"] = true
	model = updateModel(t, model, CompletionsLoadedMsg{LessonIDs: []string{"persisted"}})

	for _, lessonID := range []string{"in-session", "persisted"} {
		if !model.Lessons.Completed[lessonID] {
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
			updated, command := model.Update(AssertionsCheckedMsg{LessonID: "lesson-id", Results: test.results})
			result := updated.(State)

			if !result.Lessons.Completed["lesson-id"] {
				t.Fatal("passing assertion result did not mark lesson complete")
			}
			if command == nil {
				t.Fatal("passing assertion result returned no persistence command")
			}
			message := command()
			saved, ok := message.(CompletionSavedMsg)
			if !ok || saved.Err != nil {
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
	updated, command := model.Update(AssertionsCheckedMsg{
		LessonID: "lesson-id",
		Results:  []assertion.Result{{Passed: true}, {Passed: false}},
	})
	result := updated.(State)

	if result.Lessons.Completed["lesson-id"] {
		t.Fatal("failed assertion marked lesson complete")
	}
	if command != nil || len(store.marked) != 0 {
		t.Fatal("failed assertion attempted to persist completion")
	}
}

func TestCompletedLessonDoesNotWriteAgain(t *testing.T) {
	store := &fakeCompletionStore{}
	model := NewWithCompletionStore(store)
	model.Lessons.Completed["lesson-id"] = true
	_, command := model.Update(AssertionsCheckedMsg{LessonID: "lesson-id"})

	if command != nil || len(store.marked) != 0 {
		t.Fatal("completed lesson attempted a duplicate persistence write")
	}
}

func TestCompletionWriteFailureKeepsInMemoryCompletion(t *testing.T) {
	store := &fakeCompletionStore{markErr: errors.New("database unavailable")}
	model := NewWithCompletionStore(store)
	updated, command := model.Update(AssertionsCheckedMsg{LessonID: "lesson-id"})
	result := updated.(State)

	if !result.Lessons.Completed["lesson-id"] {
		t.Fatal("write failure removed in-memory completion")
	}
	saved := command().(CompletionSavedMsg)
	if saved.Err == nil {
		t.Fatal("completion command did not report write failure")
	}
	updated, followUp := result.Update(saved)
	if followUp != nil || !updated.(State).Lessons.Completed["lesson-id"] {
		t.Fatal("handling write failure changed completion state")
	}
}

func TestLessonExitReturnsToLessons(t *testing.T) {
	model := State{
		Nav: navigationState{Screen: lessonScreen},
		Term: terminalState{
			hasRequestedExit: true,
			Output:           "old terminal text",
			Error:            errors.New("old terminal error"),
		},
	}
	model = updateModel(t, model, TerminalExitedMsg{})
	if model.Nav.Screen != lessonsScreen {
		t.Fatalf("screen = %d after lesson terminal exit, want lessons screen", model.Nav.Screen)
	}
	if model.Term.Output != "" || model.Term.Error != nil {
		t.Fatalf("terminal state was not cleared: output=%q error=%v", model.Term.Output, model.Term.Error)
	}
}

func TestTerminalStartWithoutSessionOrErrorShowsFailure(t *testing.T) {
	model := State{Nav: navigationState{Screen: lessonScreen}}
	model = updateModel(t, model, TerminalStartedMsg{})
	if model.Term.Error == nil {
		t.Fatal("terminal error = nil, want invalid-start-result error")
	}
	if !strings.Contains(model.Term.Error.Error(), "no session and no error") {
		t.Fatalf("terminal error = %q, want invalid-start-result error", model.Term.Error)
	}
}

func TestUnexpectedLessonTerminalExitStaysOnLesson(t *testing.T) {
	model := State{Nav: navigationState{Screen: lessonScreen}}
	model = updateModel(t, model, TerminalExitedMsg{})
	if model.Nav.Screen != lessonScreen {
		t.Fatalf("screen = %d after unexpected terminal exit, want lesson screen", model.Nav.Screen)
	}
	if model.Term.Error == nil {
		t.Fatal("terminal error = nil after unexpected terminal exit")
	}

	model = updateModel(t, model, keyPress('d', "", tea.ModCtrl))
	if model.Nav.Screen != lessonsScreen {
		t.Fatalf("screen = %d after Ctrl+D acknowledgement, want lessons screen", model.Nav.Screen)
	}
}

func TestLessonBackReturnsToMainMenu(t *testing.T) {
	model := loadedState(t)
	model.Nav.Screen = lessonsScreen
	model.Nav.Selection = len(model.Lessons.Available)
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.Nav.Screen != menuScreen {
		t.Fatalf("screen = %d after selecting lesson Back, want menu screen", model.Nav.Screen)
	}
	if model.Nav.Selection != 0 {
		t.Fatalf("selection = %d after selecting lesson Back, want 0", model.Nav.Selection)
	}
}

func TestStartLearningRequestsTerminal(t *testing.T) {
	updated, command := New().Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.Nav.Screen != terminalScreen {
		t.Fatalf("screen = %d after selecting Sandbox, want terminal screen", result.Nav.Screen)
	}
	if command == nil {
		t.Fatal("selecting Sandbox returned no terminal start command")
	}
}

func TestCtrlCQuitsFromMenuAndSettings(t *testing.T) {
	for _, model := range []State{New(), {Nav: navigationState{Screen: settingsScreen}}} {
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
	model.Nav.Selection = len(menuItems) - 1
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
	if model.Viewport.Width != 100 || model.Viewport.Height != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.Viewport.Width, model.Viewport.Height)
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
	sandbox := State{Nav: navigationState{Screen: terminalScreen}, Viewport: viewportState{Width: 100, Height: 40}}
	if width, height := sandbox.terminalDimensions(); width != 96 || height != 38 {
		t.Fatalf("sandbox terminal dimensions = %dx%d, want 96x38", width, height)
	}

	lesson := State{Nav: navigationState{Screen: lessonScreen}, Viewport: viewportState{Width: 100, Height: 40}}
	if width, height := lesson.terminalDimensions(); width != 47 || height != 38 {
		t.Fatalf("lesson terminal dimensions = %dx%d, want 47x38", width, height)
	}
}

func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := State{Nav: navigationState{Screen: terminalScreen}, Term: terminalState{hasRequestedExit: true}}
	updated, command := model.Update(TerminalExitedMsg{})
	result := updated.(State)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.Nav.Screen != menuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.Nav.Screen)
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
	return State{Lessons: lessonState{Available: loaded}}
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
