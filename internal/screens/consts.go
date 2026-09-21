package screens

import "shellforge/internal/ui"

const (
	MainMenuTitle          = "Welcome to Shellforge!"
	LessonsMenuTitle       = "Choose a Lesson"
	MenuNavigationPrompt   = "Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."
	LessonNavigationPrompt = "Ctrl+P previous page · Ctrl+N next page\nF12 check progress\nCtrl+D return to lesson list"
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
