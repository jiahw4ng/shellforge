package screens

import (
	"fmt"
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
)

// LessonList renders the numbered lesson menu and its Back option.
func LessonList(available []lessons.Lesson, selection int, loadErr error) string {
	lines := []string{ui.TitleStyle.Render("Choose a lesson"), ""}
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
		lines = append(lines, selectableItem(index, selection, item))
	}

	lines = append(lines, "", selectableItem(len(available), selection, "Back"))
	lines = append(lines, "", ui.MutedStyle.Render("Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."))
	return strings.Join(lines, "\n")
}

// selectableItem adds the selection arrow and style to one menu item.
func selectableItem(index, selection int, item string) string {
	prefix := "  "
	if index == selection {
		prefix = "> "
		item = ui.SelectedStyle.Render(item)
	}
	return prefix + item
}
