package config

import "charm.land/bubbles/v2/key"

// KeyMap is the single source of truth for Shellforge's current shortcuts.
// The bindings drive input handling now and can also drive Bubbles help views.
type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	SelectItem    key.Binding
	Quit          key.Binding
	PreviousPage  key.Binding
	NextPage      key.Binding
	GuidePageUp   key.Binding
	GuidePageDown key.Binding
	ShowHint      key.Binding
	CheckProgress key.Binding
	ResetSandbox  key.Binding
	ReturnBack    key.Binding
	ToggleHelp    key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	SelectItem: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("Ctrl+C", "quit"),
	),
	PreviousPage: key.NewBinding(
		key.WithKeys("ctrl+p", "ctrl+P"),
		key.WithHelp("Ctrl+P", "previous page"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("ctrl+n", "ctrl+N"),
		key.WithHelp("Ctrl+N", "next page"),
	),
	GuidePageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("PgUp", "scroll guide up"),
	),
	GuidePageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("PgDn", "scroll guide down"),
	),
	ShowHint: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("F1", "show next hint"),
	),
	CheckProgress: key.NewBinding(
		key.WithKeys("f12"),
		key.WithHelp("F12", "check progress"),
	),
	ResetSandbox: key.NewBinding(
		key.WithKeys("ctrl+alt+r"),
		key.WithHelp("Ctrl+Alt+R", "reset sandbox"),
	),
	ReturnBack: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("Ctrl+D", "return"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("f2"),
		key.WithHelp("F2", "show help"),
	),
}

var NavigationHelpKeyBindingGroups = [][]key.Binding{
	{Keys.Up, Keys.Down, Keys.SelectItem},
	{Keys.Quit},
}

var LessonHelpKeyBindingGroups = [][]key.Binding{
	{Keys.PreviousPage, Keys.NextPage, Keys.GuidePageUp, Keys.GuidePageDown},
	{Keys.ShowHint, Keys.ResetSandbox, Keys.CheckProgress, Keys.ReturnBack},
}
