package screens

import (
	"shellforge/internal/ui"
	"strings"
)

// Settings renders the application settings menu and its latest action result.
func Settings(items []string, selection int, message string, failed bool, help string) string {
	lines := []string{ui.TitleStyle.Render("Settings"), ""}
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
