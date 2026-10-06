package app

import (
	"shellforge/internal/ui"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestNavigationHelpUsesConfiguredBindings(t *testing.T) {
	rendered := New().navigationHelp(80)
	if !strings.Contains(rendered, ui.MutedStyle.Render("┌ F3: show help")) {
		t.Fatalf("collapsed navigation help does not use the disclosure style: %q", rendered)
	}
	view := ansi.Strip(rendered)
	if view != "┌ F3: show help" {
		t.Fatalf("collapsed navigation help = %q, want compact disclosure", view)
	}

	model := New()
	model.help.ShowAll = true
	view = ansi.Strip(model.navigationHelp(80))
	for _, text := range []string{"┌ F3: hide help", "↑", "move up", "↓", "move down", "Enter", "select", "Ctrl+C", "quit"} {
		if !strings.Contains(view, text) {
			t.Errorf("navigation help does not contain %q: %q", text, view)
		}
	}
}

func TestHelpModelUsesLightPurpleStyles(t *testing.T) {
	styles := New().help.Styles
	for name, style := range map[string]lipgloss.Style{
		"ellipsis":          styles.Ellipsis,
		"short key":         styles.ShortKey,
		"short description": styles.ShortDesc,
		"short separator":   styles.ShortSeparator,
		"full key":          styles.FullKey,
		"full description":  styles.FullDesc,
		"full separator":    styles.FullSeparator,
	} {
		if style.Render("help") != ui.HelpStyle.Render("help") {
			t.Errorf("%s does not use the help style", name)
		}
	}
}

func TestLessonHelpKeepsAllConfiguredBindingsVisible(t *testing.T) {
	model := New()
	model.help.ShowAll = true
	view := ansi.Strip(model.lessonHelp(80))

	for _, text := range []string{
		"┌ F3: hide help",
		"Ctrl+P", "previous page",
		"Ctrl+N", "next page",
		"PgUp", "scroll guide up",
		"PgDn", "scroll guide down",
		"F1", "show/hide hints",
		"F2", "show/hide assertions",
		"F10", "show next hint",
		"Ctrl+Alt+R", "reset sandbox",
		"F12", "check progress",
		"Ctrl+D", "return",
	} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson help does not contain %q: %q", text, view)
		}
	}
}

func TestLessonHelpKeepsRevealHintBindingWhenNoHintsRemain(t *testing.T) {
	model := New()
	model.help.ShowAll = true

	view := ansi.Strip(model.lessonHelp(80))
	for _, text := range []string{"F10", "show next hint"} {
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
