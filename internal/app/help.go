package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

func newHelpModel() help.Model {
	model := help.New()
	model.Styles.Ellipsis = ui.HelpStyle
	model.Styles.ShortKey = ui.HelpStyle
	model.Styles.ShortDesc = ui.HelpStyle
	model.Styles.ShortSeparator = ui.HelpStyle
	model.Styles.FullKey = ui.HelpStyle
	model.Styles.FullDesc = ui.HelpStyle
	model.Styles.FullSeparator = ui.HelpStyle
	return model
}

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
