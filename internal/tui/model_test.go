package tui

import (
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
	if model.selected != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.selected)
	}

	for range menuItems {
		model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	}
	if model.selected != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.selected, len(menuItems)-1)
	}
}

func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.screen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.screen)
	}
	if !strings.Contains(model.View().Content, "feature coming soon!") {
		t.Fatal("feature view does not contain coming soon text")
	}

	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))
	if model.screen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.screen)
	}
	if model.selected != 1 {
		t.Fatalf("selected = %d after back, want 1", model.selected)
	}
}

func TestStartLearningRequestsTerminal(t *testing.T) {
	model := New()
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(Model)

	if result.screen != terminalScreen {
		t.Fatalf("screen = %d after selecting Start learning, want terminal screen", result.screen)
	}
	if command == nil {
		t.Fatal("selecting Start learning returned no terminal start command")
	}
}

func TestCtrlCQuitsFromMenuAndFeature(t *testing.T) {
	for _, model := range []Model{New(), {screen: featureScreen}} {
		_, command := model.Update(keyPress('c', "", tea.ModCtrl))
		if command == nil {
			t.Fatal("Ctrl+C returned no quit command")
		}
		if _, ok := command().(tea.QuitMsg); !ok {
			t.Fatalf("Ctrl+C command returned %T, want tea.QuitMsg", command())
		}
	}
}

func TestWindowResizeSetsDimensions(t *testing.T) {
	model := updateModel(t, New(), tea.WindowSizeMsg{Width: 100, Height: 40})
	if model.width != 100 || model.height != 40 {
		t.Fatalf("dimensions = %dx%d, want 100x40", model.width, model.height)
	}
}

func TestTerminalDimensionUsesFallbackForMissingSize(t *testing.T) {
	if got := terminalDimension(0, 80); got != 80 {
		t.Fatalf("terminalDimension(0, 80) = %d, want 80", got)
	}
	if got := terminalDimension(120, 80); got != 120 {
		t.Fatalf("terminalDimension(120, 80) = %d, want 120", got)
	}
}

func TestTerminalExitReturnsToMenu(t *testing.T) {
	model := Model{screen: terminalScreen}
	updated, command := model.Update(terminalExitedMsg{})
	result := updated.(Model)

	if command != nil {
		t.Fatal("terminal exit returned an unexpected command")
	}
	if result.screen != menuScreen {
		t.Fatalf("screen = %d after terminal exit, want menu screen", result.screen)
	}
}

func keyPress(code rune, text string, modifiers ...tea.KeyMod) tea.KeyPressMsg {
	var mod tea.KeyMod
	for _, modifier := range modifiers {
		mod |= modifier
	}
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func updateModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result
}
