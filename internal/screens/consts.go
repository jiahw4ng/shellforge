package screens

import "shellforge/internal/ui"

const (
	MainMenuTitle    = "Welcome to Shellforge!"
	LessonsMenuTitle = "Choose a Lesson"
)

// selectableItem adds the selection arrow and style to one menu item.
func selectableItem(index, selection int, item string) string {
	prefix := "  "
	if index == selection {
		prefix = "> "
		item = ui.SelectedStyle.Render(item)
	}
	return prefix + item
}

const lessonPaneGap = 1
