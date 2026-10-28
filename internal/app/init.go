package app

import (
	tea "charm.land/bubbletea/v2"
)

// Init starts loading lesson material, persisted completion state, and learner preferences.
func (s State) Init() tea.Cmd {
	// load lessons in the background
	commands := []tea.Cmd{loadLessons()}
	if s.lessons.store != nil {
		// if the persisted lesson completion store is available,
		// load the completion state in the background as well
		commands = append(commands, loadCompletedLessons(s.lessons.store))
	}
	if s.preferences.store != nil {
		// if the persisted preferences store is available,
		// load the cursor style preference in the background as well
		commands = append(commands, loadCursorStyle(s.preferences.store))
	}
	// load stuff in the background concurrently
	return tea.Batch(commands...)
}
