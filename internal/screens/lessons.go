package screens

import (
	"fmt"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
)

// LessonList renders the numbered lesson menu and its Back option.
func LessonList(available []lessons.Lesson, completed map[string]bool, selection int, loadErr error) string {
	lines := []string{ui.TitleStyle.Render(LessonsMenuTitle), ""}
	if loadErr != nil {
		lines = append(lines,
			ui.MutedStyle.Render("Lessons could not be loaded: "+loadErr.Error()),
			"",
			selectableItem(0, selection, "Back"),
		)
		return strings.Join(lines, "\n")
	}
	if len(available) == 0 {
		lines = append(lines, ui.MutedStyle.Render("Loading lessons..."))
		return strings.Join(lines, "\n")
	}

	for index, lesson := range available {
		item := fmt.Sprintf("%d. %s", lesson.Number, lesson.Title)
		if completed[lesson.ID] {
			item = "✓ " + item
		}
		lines = append(lines, selectableItem(index, selection, item))
		if index == selection && lesson.Description != "" {
			lines = append(lines, ui.MutedStyle.Render("     └── "+lesson.Description))
		}
	}

	lines = append(lines, "", selectableItem(len(available), selection, "Back"))
	lines = append(lines, "", ui.MutedStyle.Render(MenuNavigationPrompt))
	return strings.Join(lines, "\n")
}
