package tui

import (
	"shellforge/internal/lessons"

	tea "charm.land/bubbletea/v2"
)

// Init starts parsing the embedded lesson files while Bubble Tea starts the UI.
func (State) Init() tea.Cmd {
	return func() tea.Msg {
		loaded, err := lessons.Load()
		return lessonsLoadedMsg{lessons: loaded, err: err}
	}
}
