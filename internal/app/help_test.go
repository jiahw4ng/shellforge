package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNavigationHelpUsesConfiguredBindings(t *testing.T) {
	view := ansi.Strip(New().navigationHelp(80))
	if view != "> F2 show help" {
		t.Fatalf("collapsed navigation help = %q, want compact disclosure", view)
	}

	model := New()
	model.help.ShowAll = true
	view = ansi.Strip(model.navigationHelp(80))
	for _, text := range []string{"v F2 hide help", "↑", "move up", "↓", "move down", "Enter", "select", "Ctrl+C", "quit"} {
		if !strings.Contains(view, text) {
			t.Errorf("navigation help does not contain %q: %q", text, view)
		}
	}
}

func TestLessonHelpKeepsAllConfiguredBindingsVisible(t *testing.T) {
	model := New()
	model.help.ShowAll = true
	view := ansi.Strip(model.lessonHelp(80))

	for _, text := range []string{
		"v F2 hide help",
		"Ctrl+P", "previous page",
		"Ctrl+N", "next page",
		"PgUp", "scroll guide up",
		"PgDn", "scroll guide down",
		"F1", "show next hint",
		"Ctrl+Alt+R", "reset sandbox",
		"F12", "check progress",
		"Ctrl+D", "return",
	} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson help does not contain %q: %q", text, view)
		}
	}
}

func TestLessonHelpKeepsHintBindingWhenNoHintsRemain(t *testing.T) {
	model := New()
	model.help.ShowAll = true

	view := ansi.Strip(model.lessonHelp(80))
	for _, text := range []string{"F1", "show next hint"} {
		if !strings.Contains(view, text) {
			t.Fatalf("lesson help does not retain %q: %q", text, view)
		}
	}
}

func TestExpandedHelpTruncatesToLessonPaneWidth(t *testing.T) {
	model := New()
	model.help.ShowAll = true

	if view := ansi.Strip(model.lessonHelp(20)); !strings.Contains(view, "…") {
		t.Fatalf("narrow lesson help = %q, want truncated expanded help", view)
	}
}
