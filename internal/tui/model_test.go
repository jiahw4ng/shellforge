package tui

import (
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
	if model.selectedOption != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.selectedOption)
	}

	for range menuItems {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	if model.selectedOption != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.selectedOption, len(menuItems)-1)
	}
}

// TestEnterShowsFeatureAndBackRetainsSelection verifies placeholder navigation
// returns to the same menu item the learner selected.
func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.currentScreen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.currentScreen)
	}
	if !strings.Contains(model.View().Content, "feature coming soon!") {
		t.Fatal("feature view does not contain coming soon text")
	}

	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.currentScreen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.currentScreen)
	}
	if model.selectedOption != 1 {
		t.Fatalf("selected = %d after back, want 1", model.selectedOption)
	}
}

// TestStartLearningRequestsTerminal verifies the first menu item starts the
// asynchronous terminal-launch command.
func TestStartLearningRequestsTerminal(t *testing.T) {
	model := New()
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(Model)

	if result.currentScreen != terminalScreen {
		t.Fatalf("screen = %d after selecting Start learning, want terminal screen", result.currentScreen)
	}
	if command == nil {
		t.Fatal("selecting Start learning returned no terminal start command")
	}
}

// TestCtrlCQuitsFromMenuAndFeature checks Ctrl+C exits non-terminal screens.
func TestCtrlCQuitsFromMenuAndFeature(t *testing.T) {
	for _, model := range []Model{New(), {currentScreen: featureScreen}} {
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
	if model.termWidth != 100 || model.termHeight != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.termWidth, model.termHeight)
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
	model := Model{currentScreen: terminalScreen}
	updated, command := model.Update(terminalExitedMsg{})
	result := updated.(Model)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.currentScreen != menuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.currentScreen)
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
func updateModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result
}
