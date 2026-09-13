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

func TestStartLearningRequestsPTY(t *testing.T) {
	model := New()
	updated, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(Model)

	if result.screen != ptyScreen {
		t.Fatalf("screen = %d after selecting Start learning, want PTY screen", result.screen)
	}
	if command == nil {
		t.Fatal("selecting Start learning returned no PTY start command")
	}
}

func TestCtrlCQuitsFromBothScreens(t *testing.T) {
	for _, model := range []Model{New(), {screen: featureScreen}} {
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

func TestPTYInputTranslatesTerminalKeys(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
		want string
	}{
		{name: "runes", key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ls")}, want: "ls"},
		{name: "enter", key: tea.KeyMsg{Type: tea.KeyEnter}, want: "\r"},
		{name: "space", key: tea.KeyMsg{Type: tea.KeySpace}, want: " "},
		{name: "ctrl c", key: tea.KeyMsg{Type: tea.KeyCtrlC}, want: "\x03"},
		{name: "ctrl d", key: tea.KeyMsg{Type: tea.KeyCtrlD}, want: "\x04"},
		{name: "up", key: tea.KeyMsg{Type: tea.KeyUp}, want: "\x1b[A"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := string(ptyInput(test.key)); got != test.want {
				t.Fatalf("ptyInput() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestReadablePTYOutputRemovesTerminalControlSequences(t *testing.T) {
	raw := "\x1b[?2004hshellforge$ pwd\r\n/tmp/shellforge\r\n\x1b[?2004l"
	got := readablePTYOutput(raw)

	if strings.Contains(got, "\x1b") {
		t.Fatalf("readablePTYOutput() retained escape sequence in %q", got)
	}
	if got != "shellforge$ pwd\n/tmp/shellforge\n" {
		t.Fatalf("readablePTYOutput() = %q", got)
	}
}

func TestReadablePTYOutputJoinsSplitCRLFWithoutBlankLines(t *testing.T) {
	first := readablePTYOutput("pwd\r")
	second := readablePTYOutput("\n\r/tmp/shellforge\r\n")

	if first+second != "pwd\n/tmp/shellforge\n" {
		t.Fatalf("split output = %q", first+second)
	}
}

func TestAppendPTYOutputClearsTranscriptForClearCommand(t *testing.T) {
	model := Model{ptyOutput: "old transcript"}
	model.appendPTYOutput(ptyOutputMsg{
		output: "shellforge$ clear\r\n\x1b[H\x1b[2J\x1b[3Jshellforge$ ",
	})

	if model.ptyOutput != "shellforge$ " {
		t.Fatalf("PTY output after clear = %q, want prompt only", model.ptyOutput)
	}
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
