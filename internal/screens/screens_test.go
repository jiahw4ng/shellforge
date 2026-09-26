package screens

import (
	"shellforge/internal/assertion"
	"shellforge/internal/lessons"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestMainMenuShowsItems(t *testing.T) {
	items := []string{"Sandbox", "Choose lesson", "Settings", "Exit"}
	view := ansi.Strip(MainMenu(items, 0))
	for _, text := range append([]string{"Welcome to Shellforge!"}, items...) {
		if !strings.Contains(view, text) {
			t.Errorf("main menu does not contain %q", text)
		}
	}
}

func TestLessonListShowsLoadedLessonsAndBack(t *testing.T) {
	available := []lessons.Lesson{
		{ID: "01-navigation", Number: 1, Title: "Getting around", Description: "Explore a project."},
		{ID: "02-files", Number: 2, Title: "Files and directories", Description: "Create and organize files."},
	}
	view := LessonList(available, map[string]bool{"01-navigation": true}, 1, nil)
	for _, text := range []string{"1. Getting around", "Back"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson list does not contain %q", text)
		}
	}
	plainView := ansi.Strip(view)
	if !strings.Contains(plainView, "✓ 1. Getting around") {
		t.Error("lesson list does not mark the completed lesson")
	}
	if strings.Contains(plainView, "✓ 2. Files and directories") {
		t.Error("lesson list marks an incomplete lesson")
	}
	if !strings.Contains(plainView, "     └── Create and organize files.") {
		t.Error("lesson list does not contain the selected lesson's indented description")
	}
	if strings.Contains(view, "Explore a project.") {
		t.Error("lesson list contains an unselected lesson's description")
	}
}

func TestLessonUsesHalfWidthTerminal(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	view := Lesson(LessonRenderParams{
		Lesson:          &lesson,
		Page:            &page,
		PageIndex:       0,
		TerminalContent: "shellforge$ ",
		TerminalError:   nil,
		Width:           100,
		Height:          24,
	})
	plainView := ansi.Strip(view)
	for _, text := range []string{"Lesson 1: Getting around", "Page 1 of 1: pwd", "Your task", "shellforge$ ", "│"} {
		if !strings.Contains(plainView, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}

	left, right := lessonPaneWidths(100)
	if left+right+lessonPaneGap != 100 || right != 49 {
		t.Fatalf("lesson pane widths = %d and %d, want 50 and 49", left, right)
	}
}

func TestLessonShowsAssertionResults(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	view := Lesson(LessonRenderParams{
		Lesson: &lesson,
		Page:   &page,
		AssertionResults: []assertion.Result{
			{Passed: true, Message: "Directory exists."},
			{Passed: false, Message: "File does not exist."},
		},
		AssertionsChecked: true,
		Width:             100,
		Height:            24,
	})
	plainView := ansi.Strip(view)
	for _, text := range []string{"Directory exists.", "File does not exist.", "Ctrl+Alt+R reset sandbox", "F12 check progress"} {
		if !strings.Contains(plainView, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}
}

func TestLessonHidesProgressStatusUntilChecked(t *testing.T) {
	page := lessons.Page{Title: "Introduction", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}

	beforeCheck := ansi.Strip(Lesson(LessonRenderParams{
		Lesson: &lesson,
		Page:   &page,
		Width:  100,
		Height: 24,
	}))
	if strings.Contains(beforeCheck, "There are no progress checks") {
		t.Error("lesson shows the no-progress-checks status before F12")
	}

	afterCheck := ansi.Strip(Lesson(LessonRenderParams{
		Lesson:            &lesson,
		Page:              &page,
		AssertionsChecked: true,
		Width:             100,
		Height:            24,
	}))
	if !strings.Contains(afterCheck, "There are no progress checks for this page.") {
		t.Error("lesson does not show the no-progress-checks status after F12")
	}
}

func TestSettingsShowsResetAndResult(t *testing.T) {
	view := ansi.Strip(Settings([]string{"Reset lesson progress", "Back"}, 0, "Lesson progress has been reset.", false))
	for _, text := range []string{"Settings", "Reset lesson progress", "Back", "Lesson progress has been reset."} {
		if !strings.Contains(view, text) {
			t.Errorf("settings view does not contain %q", text)
		}
	}
}

func TestResetConfirmationDefaultsToNo(t *testing.T) {
	view := ansi.Strip(ResetConfirmation(
		[]string{"No, go back", "Yes, reset lesson progress"},
		0,
		false,
	))
	for _, text := range []string{"Reset lesson progress?", "cannot be undone", "> No, go back", "  Yes, reset lesson progress"} {
		if !strings.Contains(view, text) {
			t.Errorf("reset confirmation does not contain %q", text)
		}
	}
}
