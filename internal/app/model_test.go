package app

import (
	"errors"
	"fmt"
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

func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.Nav.Screen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.Nav.Screen)
	}
	if !strings.Contains(model.View().Content, "Feature coming soon!") {
		t.Fatal("feature view does not contain coming soon text")
	}

	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.Nav.Screen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.Nav.Screen)
	}
	if model.Nav.Selection != 2 {
		t.Fatalf("selected = %d after back, want 2", model.Nav.Selection)
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

func TestCtrlCQuitsFromMenuAndFeature(t *testing.T) {
	for _, model := range []State{New(), {Nav: navigationState{Screen: featureScreen}}} {
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
