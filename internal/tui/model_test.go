package tui

import (
	"fmt"
	"shellforge/internal/ui"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestInitialViewShowsWelcomeAndMenu checks the first rendered screen contains
// the greeting and every selectable menu option.
func TestInitialViewShowsWelcomeAndMenu(t *testing.T) {
	view := New().View().Content
	for _, text := range append([]string{"Welcome to Shellforge!"}, menuItems...) {
		if !strings.Contains(view, text) {
			t.Errorf("initial view does not contain %q", text)
		}
	}
}

// TestMenuNavigationStopsAtBounds ensures the selection cannot move beyond the
// first or last menu option.
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

// TestEnterShowsFeatureAndBackRetainsSelection verifies placeholder navigation
// returns to the same menu item the learner selected.
func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.CurrentScreen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.CurrentScreen)
	}
	if !strings.Contains(model.View().Content, "feature coming soon!") {
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

// TestChooseLessonShowsTenNumberedLessons confirms the lesson menu contains
// the requested progression and a selectable Back option.
func TestChooseLessonShowsTenNumberedLessons(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.CurrentScreen != lessonsScreen {
		t.Fatalf("screen = %d after choosing lessons, want lessons screen", model.CurrentScreen)
	}
	view := model.View().Content
	for index, lesson := range lessonItems {
		want := fmt.Sprintf("%d. %s", index+1, lesson)
		if !strings.Contains(view, want) {
			t.Errorf("lesson view does not contain %q", want)
		}
	}
	if !strings.Contains(view, "Back") {
		t.Error("lesson view does not contain Back")
	}
}

// TestLessonStartsSandbox verifies selecting a lesson opens the split lesson
// screen and requests a sandbox terminal session.
func TestLessonStartsSandbox(t *testing.T) {
	model := State{CurrentScreen: lessonsScreen}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.CurrentScreen != lessonScreen {
		t.Fatalf("screen = %d after selecting a lesson, want lesson screen", result.CurrentScreen)
	}
	if command == nil {
		t.Fatal("selecting a lesson returned no terminal start command")
	}
}

// TestLessonExitReturnsToLessons verifies Bash exiting from a split lesson
// closes the sandbox and returns the learner to the lesson list.
func TestLessonExitReturnsToLessons(t *testing.T) {
	model := State{CurrentScreen: lessonScreen}
	model = updateModel(t, model, terminalExitedMsg{})
	if model.CurrentScreen != lessonsScreen {
		t.Fatalf("screen = %d after lesson terminal exit, want lessons screen", model.CurrentScreen)
	}
}

// TestLessonViewUsesHalfWidthTerminal verifies the lesson layout reserves a
// right-side terminal pane and shows the selected lesson text on the left.
func TestLessonViewUsesHalfWidthTerminal(t *testing.T) {
	view := displayLessonView(0, "shellforge$ ", nil, 100, 24)
	for _, text := range []string{"Lesson 1: Getting around", "feature coming soon!", "shellforge$ ", "│"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}
	left, right := lessonPaneWidths(100)
	if left+right+lessonPaneGap != 100 || right != 49 {
		t.Fatalf("lesson pane widths = %d and %d, want 50 and 49", left, right)
	}
}

// TestLessonBackReturnsToMainMenu verifies the final Back option leaves the
// lesson menu without opening the placeholder screen.
func TestLessonBackReturnsToMainMenu(t *testing.T) {
	model := State{CurrentScreen: lessonsScreen, SelectedLesson: len(lessonItems)}
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.CurrentScreen != menuScreen {
		t.Fatalf("screen = %d after selecting lesson Back, want menu screen", model.CurrentScreen)
	}
}

// TestStartLearningRequestsTerminal verifies the first menu item starts the
// asynchronous terminal-launch command.
func TestStartLearningRequestsTerminal(t *testing.T) {
	model := New()
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.CurrentScreen != terminalScreen {
		t.Fatalf("screen = %d after selecting Sandbox, want terminal screen", result.CurrentScreen)
	}
	if command == nil {
		t.Fatal("selecting Sandbox returned no terminal start command")
	}
}

// TestCtrlCQuitsFromMenuAndFeature checks Ctrl+C exits non-terminal screens.
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

// TestWindowResizeSetsDimensions confirms the UI remembers its latest size.
func TestWindowResizeSetsDimensions(t *testing.T) {
	model := updateModel(t, New(), tea.WindowSizeMsg{Width: 100, Height: 40})
	if model.TermWidth != 100 || model.TermHeight != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.TermWidth, model.TermHeight)
	}
}

// TestApplicationFrameDrawsWhiteBorder verifies every sized application view
// is wrapped in the expected terminal border characters.
func TestApplicationFrameDrawsWhiteBorder(t *testing.T) {
	view := ui.WithDisplayApplicationFrame("Shellforge", 30, 8)
	for _, border := range []string{"┌", "┐", "└", "┘"} {
		if !strings.Contains(view, border) {
			t.Errorf("application frame does not contain %q", border)
		}
	}
}

// TestTerminalDimensionsStayInsideFrame verifies full-screen and lesson
// terminals never overwrite the application's outer border.
func TestTerminalDimensionsStayInsideFrame(t *testing.T) {
	sandbox := State{CurrentScreen: terminalScreen, TermState: TermState{TermWidth: 100, TermHeight: 40}}
	if width, height := sandbox.terminalDimensions(); width != 96 || height != 38 {
		t.Fatalf("sandbox terminal dimensions = %dx%d, want 96x38", width, height)
	}

	lesson := State{CurrentScreen: lessonScreen, TermState: TermState{TermWidth: 100, TermHeight: 40}}
	if width, height := lesson.terminalDimensions(); width != 47 || height != 38 {
		t.Fatalf("lesson terminal dimensions = %dx%d, want 47x38", width, height)
	}
}

// TestTerminalDimensionUsesFallbackForMissingSize protects startup before the
// first terminal-size event arrives.
func TestTerminalDimensionUsesFallbackForMissingSize(t *testing.T) {
	if got := terminalDimension(0, 80); got != 80 {
		t.Fatalf("terminalDimension(0, 80) = %d, want 80", got)
	}
	if got := terminalDimension(120, 80); got != 120 {
		t.Fatalf("terminalDimension(120, 80) = %d, want 120", got)
	}
}

// TestTerminalExitReturnsToMenu checks a closed Bash session returns learners
// to the main menu.
func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := State{CurrentScreen: terminalScreen}
	updated, command := model.Update(terminalExitedMsg{})
	result := updated.(State)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.CurrentScreen != menuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.CurrentScreen)
	}
}

// keyPress builds readable Bubble Tea key events for the menu tests.
func keyPress(code rune, text string, modifiers ...tea.KeyMod) tea.KeyPressMsg {
	var mod tea.KeyMod
	for _, modifier := range modifiers {
		mod |= modifier
	}
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

// updateModel applies one message and asserts the outer model keeps its type.
func updateModel(t *testing.T, model State, message tea.Msg) State {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(State)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result
}
