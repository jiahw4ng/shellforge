package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/key"
)

// navigationHelp renders toggleable shortcuts shared by non-terminal menu screens.
func (s State) navigationHelp(width int) string {
	return s.renderHelp(width, config.NavigationHelpKeyBindingGroups)
}

// lessonHelp renders toggleable shortcuts for the current lesson state.
func (s State) lessonHelp(width int) string {
	return s.renderHelp(width, config.LessonHelpKeyBindingGroups)
}

func (s State) renderHelp(width int, groups [][]key.Binding) string {
	helpModel := s.help
	helpModel.SetWidth(width)
	body := helpModel.FullHelpView(groups)
	return ui.Disclosure(config.Keys.ToggleHelp, s.help.ShowAll, "show help", "hide help", body)
}
