package screens

import "shellforge/internal/ui"

const (
	mainMenuTitle    = "Welcome to Shellforge!"
	lessonsMenuTitle = "Choose a Lesson"
)

// selectableItem adds the selection arrow and style to one menu item.
func selectableItem(index, selection int, item string) string {
	prefix := "  "
	if index == selection {
		prefix = "▶ "
		item = ui.SelectedStyle.Render(item)
	}
	return prefix + item
}

const lessonPaneGap = 1
