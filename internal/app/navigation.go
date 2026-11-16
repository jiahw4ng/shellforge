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
//
// A visit to a child screen (see screenBehaviour.parent) keeps the parent's
// state: opening the child skips the parent's onDeparture, and returning to
// the parent skips its onArrival.
func (s *State) changeScreen(nextState navigationState) tea.Cmd {
	fromScreen := s.nav.screen
	toScreen := nextState.screen

	// Leave the old screen while it is still active, unless opening its child.
	var departureWork tea.Cmd
	if !isChildOf(toScreen, fromScreen) {
		departureWork = s.uponDepartureFromScreen(toScreen)
	}

	s.nav = nextState

	// Arrive at the new screen, unless returning to it from its child.
	var arrivalWork tea.Cmd
	if !isChildOf(fromScreen, toScreen) {
		arrivalWork = s.uponArrivalAtNewScreen(fromScreen)
	}

	return tea.Batch(departureWork, arrivalWork)
}

// uponDepartureFromScreen does the work the current screen needs when it is
// left for the screen given.
func (s *State) uponDepartureFromScreen(to config.Screen) tea.Cmd {
	if depart := s.activeScreenBehaviour().onDeparture; depart != nil {
		return depart(s, to)
	}
	return nil
}

// uponArrivalAtNewScreen starts the work the current screen needs when it is entered from
// another screen. Moving the selection never calls it, because only changeScreen
// changes screen.
func (s *State) uponArrivalAtNewScreen(from config.Screen) tea.Cmd {
	if arrive := s.activeScreenBehaviour().onArrival; arrive != nil {
		return arrive(s, from)
	}
	return nil
}

// selectableItemCount is how many items Up and Down move between on the
// active screen.
func (s State) selectableItemCount() int {
	if list := s.activeScreenBehaviour().list; list != nil {
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
	if list := s.activeScreenBehaviour().list; list != nil {
		return list.onSelectItem(s)
	}
	return nil
}
