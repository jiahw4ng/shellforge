package screens

import "shellforge/internal/ui"

const (
	MainMenuTitle    = "Welcome to Shellforge!"
	LessonsMenuTitle = "Choose a Lesson"

	unselectedItemPrefix     = "  "
	selectedItemPrefix       = "▶ "
	lessonDescriptionPrefix  = "     └── "
	hintPrompt               = "Press F10 to show a hint."
	paginatorDot             = "•"
	moreAboveLabel           = "↑ More above"
	moreBelowLabel           = "↓ More below"
	scrollStatusSeparator    = " · "
	progressCheckPrompt      = "Press F12 to check progress."
	progressCheckingMessage  = "Checking Progress..."
	allChecksPassedMessage   = "All checks passed!"
	checksUnavailableMessage = "Unable to run some checks!"
	passedCheckPrefix        = "✓ "
	pendingCheckPrefix       = "○ "
	failedCheckPrefix        = "✗ "
	noProgressChecksMessage  = "There are no progress checks for this page."
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
