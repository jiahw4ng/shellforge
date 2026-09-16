package tui

import (
	"fmt"
	lessondefs "shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
)

// displayLessonsView renders the numbered lesson menu and its Back option.
func displayLessonsView(lessons []lessondefs.Lesson, selection int, loadErr error) string {
	lines := []string{ui.TitleStyle.Render("Choose a lesson"), ""}
	if loadErr != nil {
		lines = append(lines,
			ui.MutedStyle.Render("Lessons could not be loaded: "+loadErr.Error()),
			"",
			selectableLessonItem(0, selection, "Back"),
		)
		return strings.Join(lines, "\n")
	}
	if len(lessons) == 0 {
		lines = append(lines, ui.MutedStyle.Render("Loading lessons..."))
		return strings.Join(lines, "\n")
	}

	for index, lesson := range lessons {
		item := fmt.Sprintf("%d. %s", lesson.Number, lesson.Title)
		lines = append(lines, selectableLessonItem(index, selection, item))
	}

	lines = append(lines, "", selectableLessonItem(len(lessons), selection, "Back"))
	lines = append(lines, "", ui.MutedStyle.Render("Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."))
	return strings.Join(lines, "\n")
}

// selectableLessonItem adds the selection arrow and style to one lesson-menu item.
func selectableLessonItem(index, selection int, item string) string {
	prefix := "  "
	if index == selection {
		prefix = "> "
		item = ui.SelectedStyle.Render(item)
	}
	return prefix + item
}
