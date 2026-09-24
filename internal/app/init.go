package app

import (
	tea "charm.land/bubbletea/v2"
)

// Init starts loading lesson material and persisted completion state.
func (m State) Init() tea.Cmd {
	// load lessons in the background
	commands := []tea.Cmd{loadLessons()}
	if m.lessons.store != nil {
		// if the persisted lesson completion store is available,
		// load the completion state in the background as well
		commands = append(commands, loadCompletedLessons(m.lessons.store))
	}
	// load stuff in the background concurrently
	return tea.Batch(commands...)
}
