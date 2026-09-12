package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialViewShowsWelcomeAndMenu(t *testing.T) {
	view := New().View()

	for _, text := range append([]string{"Welcome to Shellforge!"}, menuItems...) {
		if !strings.Contains(view, text) {
			t.Errorf("initial view does not contain %q", text)
		}
	}
}

func TestMenuNavigationStopsAtBounds(t *testing.T) {
	model := New()
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyUp})
	if model.selected != 0 {
		t.Fatalf("selected = %d after moving up from first item, want 0", model.selected)
	}

	for range menuItems {
		model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyDown})
	}
	if model.selected != len(menuItems)-1 {
		t.Fatalf("selected = %d after moving down, want %d", model.selected, len(menuItems)-1)
	}
}

func TestEnterShowsFeatureAndBackRetainsSelection(t *testing.T) {
	model := New()
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyDown})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.screen != featureScreen {
		t.Fatalf("screen = %d after enter, want feature screen", model.screen)
	}
	if !strings.Contains(model.View(), "feature coming soon!") {
		t.Fatal("feature view does not contain coming soon text")
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.screen != menuScreen {
		t.Fatalf("screen = %d after back, want menu screen", model.screen)
	}
	if model.selected != 1 {
		t.Fatalf("selected = %d after back, want 1", model.selected)
	}
}

func TestCtrlCQuitsFromBothScreens(t *testing.T) {
	for _, model := range []SFModel{New(), {screen: featureScreen}} {
		_, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
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

func updateModel(t *testing.T, model SFModel, message tea.Msg) SFModel {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(SFModel)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.Model", updated)
	}
	return result
}
