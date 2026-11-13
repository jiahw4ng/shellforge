package screens

import "shellforge/internal/ui"

const (
	MainMenuTitle              = "Welcome to Shellforge!"
	LessonsMenuTitle           = "Choose a Lesson"
	sandboxTitle               = ">_ Sandboxed Bash"
	sandboxQuitTitleFormat     = ">_ Sandboxed Bash (type %s to quit)"
	startingSandbox            = "Starting sandboxed shell..."
	checkingDockerLine         = "Checking that Docker is running"
	checkingImageLine          = "Checking if sandbox image exists"
	pullingImageLine           = "Image does not exist, pulling from registry"
	startingImageFoundLine     = "Image found, starting container"
	startingAfterPullLine      = "Image pulled, starting container"
	preparingLessonLine        = "Preparing lesson"
	connectingShellLine        = "Connecting to the shell"
	terminalErrorTitle         = "Oops! There was an error."
	logPathPrefix              = "Full details: "
	unselectedItemPrefix       = "  "
	selectedItemPrefix         = "▶ "
	lessonDescriptionPrefix    = "     └── "
	hintPrompt                 = "Press F10 to show a hint."
	paginatorDot               = "•"
	moreAboveLabel             = "↑ More above"
	moreBelowLabel             = "↓ More below"
	scrollStatusSeparator      = " · "
	lessonRenderFailureMessage = "Unable to display lesson content."
	progressCheckPrompt        = "Press F12 to check progress."
	progressCheckingMessage    = "Checking Progress..."
	allChecksPassedMessage     = "All checks passed!"
	checksUnavailableMessage   = "Unable to run some checks!"
	passedCheckPrefix          = "✓ "
	pendingCheckPrefix         = "○ "
	failedCheckPrefix          = "✗ "
	noProgressChecksMessage    = "There are no progress checks for this page."
	terminalHeaderHeight       = 1
	lessonPaneGap              = 1
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
