package tui

import (
	"fmt"
	"shellforge/internal/ui"
	"strings"
)

// displayLessonsView renders the numbered lesson menu and its Back option.
func displayLessonsView(selection int) string {
	lines := []string{ui.TitleStyle.Render("Choose a lesson"), ""}
	for index, lesson := range lessonItems {
		item := fmt.Sprintf("%d. %s", index+1, lesson)
		lines = append(lines, selectableLessonItem(index, selection, item))
	}

	lines = append(lines, "", selectableLessonItem(len(lessonItems), selection, "Back"))
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
