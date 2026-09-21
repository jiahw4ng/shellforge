package app

import (
	tea "charm.land/bubbletea/v2"
)

// Init starts loading lesson material and persisted completion state.
func (m State) Init() tea.Cmd {
	commands := []tea.Cmd{loadLessons()}
	if m.Lessons.Store != nil {
		commands = append(commands, loadCompletedLessons(m.Lessons.Store))
	}
	return tea.Batch(commands...)
}
