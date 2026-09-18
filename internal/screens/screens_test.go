package screens

import (
	"shellforge/internal/lessons"
	"strings"
	"testing"
)

func TestMainMenuShowsItems(t *testing.T) {
	items := []string{"Sandbox", "Choose lesson", "Settings"}
	view := MainMenu(items, 0)
	for _, text := range append([]string{"Welcome to Shellforge!"}, items...) {
		if !strings.Contains(view, text) {
			t.Errorf("main menu does not contain %q", text)
		}
	}
}

func TestLessonListShowsLoadedLessonsAndBack(t *testing.T) {
	available := []lessons.Lesson{{Number: 1, Title: "Getting around"}}
	view := LessonList(available, 0, nil)
	for _, text := range []string{"1. Getting around", "Back"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson list does not contain %q", text)
		}
	}
}

func TestLessonUsesHalfWidthTerminal(t *testing.T) {
	page := lessons.Page{Title: "pwd", Content: "Your task"}
	lesson := lessons.Lesson{Number: 1, Title: "Getting around", Pages: []lessons.Page{page}}
	view := Lesson(lesson, page, 0, "shellforge$ ", nil, 100, 24)
	for _, text := range []string{"Lesson 1: Getting around", "Page 1 of 1: pwd", "Your task", "shellforge$ ", "│"} {
		if !strings.Contains(view, text) {
			t.Errorf("lesson view does not contain %q", text)
		}
	}

	left, right := lessonPaneWidths(100)
	if left+right+lessonPaneGap != 100 || right != 49 {
		t.Fatalf("lesson pane widths = %d and %d, want 50 and 49", left, right)
	}
}

func TestFeatureShowsBackOption(t *testing.T) {
	view := Feature()
	for _, text := range []string{"feature coming soon!", "Back"} {
		if !strings.Contains(view, text) {
			t.Errorf("feature view does not contain %q", text)
		}
	}
}
