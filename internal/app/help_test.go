package app

import (
	"shellforge/internal/lessons"
	"shellforge/internal/terminal"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNavigationHelpUsesConfiguredBindings(t *testing.T) {
	view := ansi.Strip(New().navigationHelp())
	for _, text := range []string{"↑ move up", "↓ move down", "Enter select", "Ctrl+C quit"} {
		if !strings.Contains(view, text) {
			t.Errorf("navigation help does not contain %q: %q", text, view)
		}
	}
}

func TestLessonHelpShowsOnlyAvailableBindings(t *testing.T) {
	model := State{
		lessons: lessonState{
			available: []lessons.Lesson{{Pages: []lessons.Page{{}, {}}}},
		},
		term: terminalState{session: &terminal.TermSession{}},
	}
	view := ansi.Strip(model.lessonHelp())

	for _, text := range []string{"Ctrl+N next page", "Ctrl+Alt+R reset sandbox", "F12 check progress", "Ctrl+D return"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson help does not contain %q: %q", text, view)
		}
	}
	if strings.Contains(view, "Ctrl+P previous page") {
		t.Fatalf("lesson help shows previous page on the first page: %q", view)
	}
}
