package screens

import "shellforge/internal/ui"

const (
	MainMenuTitle    = "Welcome to Shellforge!"
	LessonsMenuTitle = "Choose a Lesson"

	unselectedItemPrefix    = "  "
	selectedItemPrefix      = "▶ "
	lessonDescriptionPrefix = "     └── "
	hintPrompt              = "Press F10 to show a hint."
	paginatorDot            = "•"
	passedCheckPrefix       = "✓ "
	failedCheckPrefix       = "✗ "
)

// selectableItem adds the selection arrow and style to one menu item.
func selectableItem(index, selection int, item string) string {
	prefix := unselectedItemPrefix
	if index == selection {
		prefix = selectedItemPrefix
		item = ui.SelectedStyle.Render(item)
	}
	return prefix + item
}

const lessonPaneGap = 1
