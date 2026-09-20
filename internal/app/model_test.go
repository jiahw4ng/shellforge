package app

import (
	"errors"
	"fmt"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
	"testing"

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

func TestMenuNavigationStopsAtBounds(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyUp, ""))
	if model.SelectedOption != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.SelectedOption)
	}

	for range menuItems {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	if model.SelectedOption != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.SelectedOption, len(menuItems)-1)
	}
}

func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.CurrentScreen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.CurrentScreen)
	}
	if !strings.Contains(model.View().Content, "Feature coming soon!") {
		t.Fatal("feature view does not contain coming soon text")
	}

	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.CurrentScreen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.CurrentScreen)
	}
	if model.SelectedOption != 2 {
		t.Fatalf("selected = %d after back, want 2", model.SelectedOption)
	}
}

func TestInitLoadsEmbeddedLessons(t *testing.T) {
	model := New()
	model = updateModel(t, model, model.Init()())

	if model.LessonErr != nil {
		t.Fatalf("lesson load error = %v", model.LessonErr)
	}
	if len(model.Lessons) != 2 {
		t.Fatalf("loaded lessons = %d, want 2", len(model.Lessons))
	}
	if model.Lessons[0].ID != "01-navigation" {
		t.Fatalf("first lesson ID = %q, want 01-navigation", model.Lessons[0].ID)
	}
}

func TestChooseLessonShowsLoadedLessons(t *testing.T) {
	model := loadedState(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.CurrentScreen != lessonsScreen {
		t.Fatalf("screen = %d after choosing lessons, want lessons screen", model.CurrentScreen)
	}
	view := model.View().Content
	for _, lesson := range model.Lessons {
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
	model.CurrentScreen = lessonsScreen
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.CurrentScreen != lessonScreen {
		t.Fatalf("screen = %d after selecting a lesson, want lesson screen", result.CurrentScreen)
	}
	if command == nil {
		t.Fatal("selecting a lesson returned no terminal start command")
	}
	if !result.TerminalStarting {
		t.Fatal("terminal startup spinner was not enabled")
	}
}

func TestLessonPageNavigationUsesCtrlBrackets(t *testing.T) {
	model := State{
		CurrentScreen: lessonScreen,
		Lessons: []lessons.Lesson{{
			Pages: []lessons.Page{{Title: "pwd", Content: "first"}, {Title: "ls", Content: "second"}},
		}},
	}

	model = updateModel(t, model, keyPress(']', "]", tea.ModCtrl))
	if model.ActivePage != 1 {
		t.Fatalf("active page = %d after Ctrl+], want 1", model.ActivePage)
	}
	model = updateModel(t, model, keyPress(']', "]", tea.ModCtrl))
	if model.ActivePage != 1 {
		t.Fatalf("active page = %d beyond final page, want 1", model.ActivePage)
	}
	model = updateModel(t, model, keyPress('[', "[", tea.ModCtrl))
	if model.ActivePage != 0 {
		t.Fatalf("active page = %d after Ctrl+[, want 0", model.ActivePage)
	}
}

func TestF12DoesNotStartAssertionsWithoutATerminal(t *testing.T) {
	model := State{CurrentScreen: lessonScreen, Lessons: []lessons.Lesson{{}}}
	updated, command := model.Update(keyPress(tea.KeyF12, ""))
	result := updated.(State)

	if command != nil {
		t.Fatal("F12 without a terminal returned an assertion command")
	}
	if result.AssertionsChecking {
		t.Fatal("F12 without a terminal started an assertion check")
	}
}

func TestLessonExitReturnsToLessons(t *testing.T) {
	model := State{
		CurrentScreen:         lessonScreen,
		TerminalExitRequested: true,
		TerminalOutput:        "old terminal text",
		TerminalErr:           errors.New("old terminal error"),
	}
	model = updateModel(t, model, TerminalExitedMsg{})
	if model.CurrentScreen != lessonsScreen {
		t.Fatalf("screen = %d after lesson terminal exit, want lessons screen", model.CurrentScreen)
	}
	if model.TerminalOutput != "" || model.TerminalErr != nil {
		t.Fatalf("terminal state was not cleared: output=%q error=%v", model.TerminalOutput, model.TerminalErr)
	}
}

func TestTerminalStartWithoutSessionOrErrorShowsFailure(t *testing.T) {
	model := State{CurrentScreen: lessonScreen}
	model = updateModel(t, model, TerminalStartedMsg{})
	if model.TerminalErr == nil {
		t.Fatal("terminal error = nil, want invalid-start-result error")
	}
	if !strings.Contains(model.TerminalErr.Error(), "no session and no error") {
		t.Fatalf("terminal error = %q, want invalid-start-result error", model.TerminalErr)
	}
}

func TestUnexpectedLessonTerminalExitStaysOnLesson(t *testing.T) {
	model := State{CurrentScreen: lessonScreen}
	model = updateModel(t, model, TerminalExitedMsg{})
	if model.CurrentScreen != lessonScreen {
		t.Fatalf("screen = %d after unexpected terminal exit, want lesson screen", model.CurrentScreen)
	}
	if model.TerminalErr == nil {
		t.Fatal("terminal error = nil after unexpected terminal exit")
	}

	model = updateModel(t, model, keyPress('d', "", tea.ModCtrl))
	if model.CurrentScreen != lessonsScreen {
		t.Fatalf("screen = %d after Ctrl+D acknowledgement, want lessons screen", model.CurrentScreen)
	}
}

func TestLessonBackReturnsToMainMenu(t *testing.T) {
	model := loadedState(t)
	model.CurrentScreen = lessonsScreen
	model.SelectedLesson = len(model.Lessons)
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.CurrentScreen != menuScreen {
		t.Fatalf("screen = %d after selecting lesson Back, want menu screen", model.CurrentScreen)
	}
}

func TestStartLearningRequestsTerminal(t *testing.T) {
	updated, command := New().Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.CurrentScreen != terminalScreen {
		t.Fatalf("screen = %d after selecting Sandbox, want terminal screen", result.CurrentScreen)
	}
	if command == nil {
		t.Fatal("selecting Sandbox returned no terminal start command")
	}
}

func TestCtrlCQuitsFromMenuAndFeature(t *testing.T) {
	for _, model := range []State{New(), {CurrentScreen: featureScreen}} {
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
	model.SelectedOption = len(menuItems) - 1
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
	if model.Width != 100 || model.Height != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.Width, model.Height)
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
	sandbox := State{CurrentScreen: terminalScreen, Width: 100, Height: 40}
	if width, height := sandbox.terminalDimensions(); width != 96 || height != 38 {
		t.Fatalf("sandbox terminal dimensions = %dx%d, want 96x38", width, height)
	}

	lesson := State{CurrentScreen: lessonScreen, Width: 100, Height: 40}
	if width, height := lesson.terminalDimensions(); width != 47 || height != 38 {
		t.Fatalf("lesson terminal dimensions = %dx%d, want 47x38", width, height)
	}
}

func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := State{CurrentScreen: terminalScreen, TerminalExitRequested: true}
	updated, command := model.Update(TerminalExitedMsg{})
	result := updated.(State)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.CurrentScreen != menuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.CurrentScreen)
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
	return State{Lessons: loaded}
}
