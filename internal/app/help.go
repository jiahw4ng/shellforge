package app

import (
	"charm.land/bubbles/v2/key"
)

// helpBindings adapts the current screen's shortcuts for Bubbles' help view.
// It keeps display-only layouts separate from keyMap, which handles app input.
type helpBindings struct {
	short []key.Binding
	full  [][]key.Binding
}

func (k helpBindings) ShortHelp() []key.Binding {
	return k.short
}

func (k helpBindings) FullHelp() [][]key.Binding {
	return k.full
}

// navigationHelp renders toggleable shortcuts shared by non-terminal menu screens.
func (s State) navigationHelp(width int) string {
	return s.renderHelp(width, navigationHelpKeyBindingGroups)
}

// lessonHelp renders toggleable shortcuts for the current lesson state.
func (s State) lessonHelp(width int) string {
	return s.renderHelp(width, lessonHelpKeyBindingGroups)
}

func (s State) renderHelp(width int, groups [][]key.Binding) string {
	toggleHelp := keys.toggleHelp
	disclosure := ">"
	description := "show help"
	if s.help.ShowAll {
		disclosure = "v"
		description = "hide help"
	}
	toggleHelp.SetHelp("F2", description)

	helpModel := s.help
	helpWidth := width
	if !s.help.ShowAll {
		helpWidth = max(width-2, 1)
	}
	helpModel.SetWidth(helpWidth)
	keyMap := helpBindings{
		short: []key.Binding{toggleHelp},
		full:  groups,
	}
	if s.help.ShowAll {
		return disclosure + " " + helpModel.ShortHelpView([]key.Binding{toggleHelp}) + "\n" + helpModel.View(keyMap)
	}
	return disclosure + " " + helpModel.View(keyMap)
}
