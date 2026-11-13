package app

import (
	tea "charm.land/bubbletea/v2"
)

// Init starts loading lesson material and persisted completion state. Learner
// preferences are already loaded by NewWithStorage. When no theme has been
// chosen, it also asks the terminal for its background color.
func (s State) Init() tea.Cmd {
	// load lessons in the background
	commands := []tea.Cmd{loadLessons()}
	if !s.preferences.themeChosen {
		commands = append(commands, tea.RequestBackgroundColor)
	}
	if s.lessons.store != nil {
		// if the persisted lesson completion store is available,
		// load the completion state in the background as well
		commands = append(commands, loadCompletedLessons(s.lessons.store))
	}
	// load stuff in the background concurrently
	return tea.Batch(commands...)
}
