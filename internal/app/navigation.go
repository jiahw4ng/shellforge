package app

import (
	"shellforge/internal/config"

	tea "charm.land/bubbletea/v2"
)

// confirmItem is the confirming choice on every confirmation screen. Each one
// lists the safe choice first, at index 0, so Enter on the first item a
// confirmation opens with never deletes or resets anything.
const confirmItem = 1

// changeScreen is the only way to change screen. It runs the old screen's
// onDeparture, shows nextState.screen with nextState.selection selected (the
// first item when left out), then runs the new screen's onArrival. It returns
// the work both started, so no caller can change screen and forget either.
func (s *State) changeScreen(nextState navigationState) tea.Cmd {
	// this is the old screen
	from := s.nav.screen
	// leave it while it is still the active screen
	departure := s.uponDepartureFromScreen(nextState.screen)
	// this is the new state
	s.nav = nextState
	return tea.Batch(departure, s.uponArrivalAtNewScreen(from))
}

// uponDepartureFromScreen does the work the current screen needs when it is
// left for the screen given.
func (s *State) uponDepartureFromScreen(to config.Screen) tea.Cmd {
	if depart := s.behaviour().onDeparture; depart != nil {
		return depart(s, to)
	}
	return nil
}

// uponArrivalAtNewScreen starts the work the current screen needs when it is entered from
// another screen. Moving the selection never calls it, because only changeScreen
// changes screen.
func (s *State) uponArrivalAtNewScreen(from config.Screen) tea.Cmd {
	if arrive := s.behaviour().onArrival; arrive != nil {
		return arrive(s, from)
	}
	return nil
}

// selectableItemCount is how many items Up and Down move between on the
// active screen.
func (s State) selectableItemCount() int {
	if list := s.behaviour().list; list != nil {
		return list.itemCount(s)
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
	if list := s.behaviour().list; list != nil {
		return list.onSelectItem(s)
	}
	return nil
}
