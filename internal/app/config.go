package app

import "charm.land/bubbles/v2/key"

type screen int

const (
	menuScreen screen = iota
	lessonsScreen
	lessonScreen
	settingsScreen
	resetConfirmationScreen
	terminalScreen
)

var menuItems = []string{
	"Sandbox",
	"Choose lesson",
	"Settings",
	"Exit",
}

var settingsItems = []string{
	"Reset lesson progress",
	"Back",
}

var resetConfirmationItems = []string{
	"No, go back",
	"Yes, reset lesson progress",
}

// keyMap is the single source of truth for Shellforge's current shortcuts.
// The bindings drive input handling now and can also drive Bubbles help views.
type keyMap struct {
	up            key.Binding
	down          key.Binding
	selectItem    key.Binding
	quit          key.Binding
	previousPage  key.Binding
	nextPage      key.Binding
	checkProgress key.Binding
	resetSandbox  key.Binding
	returnBack    key.Binding
}

var keys = keyMap{
	up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	selectItem: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "select"),
	),
	quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("Ctrl+C", "quit"),
	),
	previousPage: key.NewBinding(
		key.WithKeys("ctrl+p", "ctrl+P"),
		key.WithHelp("Ctrl+P", "previous page"),
	),
	nextPage: key.NewBinding(
		key.WithKeys("ctrl+n", "ctrl+N"),
		key.WithHelp("Ctrl+N", "next page"),
	),
	checkProgress: key.NewBinding(
		key.WithKeys("f12"),
		key.WithHelp("F12", "check progress"),
	),
	resetSandbox: key.NewBinding(
		key.WithKeys("ctrl+alt+r"),
		key.WithHelp("Ctrl+Alt+R", "reset sandbox"),
	),
	returnBack: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("Ctrl+D", "return"),
	),
}
