package screens

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"
	"strings"
)

// Settings renders the application settings menu and its latest action result.
func Settings(items []string, selection int, message string, failed bool, help string) string {
	lines := []string{ui.TitleStyle.Render(config.SettingsMenuLabel), ""}
	for index, item := range items {
		if index == len(items)-1 {
			lines = append(lines, "")
		}
		lines = append(lines, selectableItem(index, selection, item))
	}
	if message != "" {
		style := ui.SuccessStyle
		if failed {
			style = ui.FailureStyle
		}
		lines = append(lines, "", style.Render(message))
	}
	lines = append(lines, "", help)
	return strings.Join(lines, "\n")
}

// ResetConfirmation asks the learner to confirm deletion of completion data.
func ResetConfirmation(items []string, selection int, resetting bool, help string) string {
	lines := []string{
		ui.TitleStyle.Render("Reset lesson progress?"),
		"",
		"This removes every lesson completion checkmark.",
		"This action cannot be undone.",
		"",
	}
	for index, item := range items {
		lines = append(lines, selectableItem(index, selection, item))
	}
	if resetting {
		lines = append(lines, "", ui.MutedStyle.Render("Resetting lesson progress..."))
	}
	if help != "" {
		lines = append(lines, "", help)
	}
	return strings.Join(lines, "\n")
}

// ResetSettingsConfirmation asks the learner to confirm restoring the original
// cursor style and theme.
func ResetSettingsConfirmation(items []string, selection int, help string) string {
	lines := []string{
		ui.TitleStyle.Render("Reset settings?"),
		"",
		"This restores the cursor style and theme to their defaults.",
		"Lesson progress is not affected.",
		"",
	}
	for index, item := range items {
		lines = append(lines, selectableItem(index, selection, item))
	}
	if help != "" {
		lines = append(lines, "", help)
	}
	return strings.Join(lines, "\n")
}

// CursorStyle lists the cursor style choices, marks the current one, and ends with Back.
func CursorStyle(items []string, current, selection int, help string) string {
	return choiceScreen(config.SettingsItemCursorStyle, items, current, selection,
		"The cursor style applies in the sandbox and lesson terminals.", help)
}

// Theme lists the color theme choices, marks the current one, and ends with Back.
func Theme(items []string, current, selection int, help string) string {
	return choiceScreen(config.SettingsItemTheme, items, current, selection,
		"The theme applies to menus and lesson text, not to the sandbox prompt.", help)
}

// choiceScreen renders a single-choice settings list: the current choice is
// checked and Back is always the last selectable item.
func choiceScreen(title string, items []string, current, selection int, note, help string) string {
	lines := []string{ui.TitleStyle.Render(title), ""}
	for index, item := range items {
		if index == current {
			item = passedCheckPrefix + item
		}
		lines = append(lines, selectableItem(index, selection, item))
	}
	lines = append(lines,
		"",
		selectableItem(len(items), selection, config.BackMenuLabel),
		"",
		ui.MutedStyle.Render(note),
		"",
		help,
	)
	return strings.Join(lines, "\n")
}
