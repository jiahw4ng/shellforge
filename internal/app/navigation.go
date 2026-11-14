package app

import (
	"shellforge/internal/config"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// handleNavigation moves the selection or acts on Enter for non-terminal
// screens. It returns tea.Quit when the learner quits, and otherwise the work
// started by entering another screen, if any.
func (s *State) handleNavigation(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, config.Keys.Quit):
		return tea.Quit
	case key.Matches(msg, config.Keys.Up):
		s.moveSelection(-1)
	case key.Matches(msg, config.Keys.Down):
		s.moveSelection(1)
	case key.Matches(msg, config.Keys.SelectItem):
		return s.handleEnter()
	}
	return nil
}

// confirmItem is the confirming choice on every confirmation screen. Each one
// lists the safe choice first, at index 0, so Enter on the first item a
// confirmation opens with never deletes or resets anything.
const confirmItem = 1

// navigate is the only way to change screen. It shows to.screen with
// to.selection selected (the first item when left out) and returns the work
// that entering it starts, so no caller can change screen and forget that work.
func (s *State) navigate(nextState navigationState) tea.Cmd {
	// this is the old screen
	from := s.nav.screen
	// this is the new state
	s.nav = nextState
	return s.arrive(from)
}

// arrive starts the work the current screen needs when it is entered from
// another screen. Moving the selection never calls it, because only navigate
// changes screen.
func (s *State) arrive(from config.Screen) tea.Cmd {
	if arrive := s.behaviour().arrive; arrive != nil {
		return arrive(s, from)
	}
	return nil
}

// selectableItemCount is how many items Up and Down move between on the
// active screen.
func (s State) selectableItemCount() int {
	if itemCount := s.behaviour().itemCount; itemCount != nil {
		return itemCount(s)
	}
	return 0
}

func (s *State) moveSelection(delta int) {
	next := s.nav.selection + delta
	if next >= 0 && next < s.selectableItemCount() {
		s.nav.selection = next
	}
}

// handleEnter acts on the selected item on the active screen and returns the
// work started by any screen it enters.
func (s *State) handleEnter() tea.Cmd {
	if selectItem := s.behaviour().selectItem; selectItem != nil {
		return selectItem(s)
	}
	return nil
}
