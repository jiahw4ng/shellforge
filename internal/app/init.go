package app

import (
	tea "charm.land/bubbletea/v2"
)

// Init starts loading lesson material and persisted completion state.
func (s State) Init() tea.Cmd {
	commands := []tea.Cmd{loadLessons()}
	if s.lessons.store != nil {
		commands = append(commands, loadCompletedLessons(s.lessons.store))
	}
	return tea.Batch(commands...)
}
