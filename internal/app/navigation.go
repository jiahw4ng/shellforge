package app

import (
	"fmt"
	"shellforge/internal/config"
	"slices"

	tea "charm.land/bubbletea/v2"
)

// confirmItem is the confirming choice on every confirmation screen. Each one
// lists the safe choice first, at index 0, so Enter on the first item a
// confirmation opens with never deletes or resets anything.
const confirmItem = 1

// openScreen and goBack are the only ways to change screen. Screens form a
// tree (see screenBehaviour.parent) and navStack is the path from the root to
// the active screen, so a screen is open from openScreen until goBack closes
// it, and its parent stays open underneath meanwhile.

// openScreen opens child.screen, a child of the active screen, with
// child.selection selected (the first item when left out), and returns the
// work its onArrival starts. The active screen stays open underneath with its
// selection, so goBack returns to it as it was. Opening a screen that is not a
// child of the active one is a programming error and panics.
func (s *State) openScreen(child navigationState) tea.Cmd {
	from := s.nav.screen
	if parent := behaviourOf(child.screen).parent; parent == nil || *parent != from {
		panic(fmt.Sprintf("screen %d opened from screen %d, which is not its parent", child.screen, from))
	}
	// push the parent into the existing stack
	s.navStack = append(slices.Clip(s.navStack), s.nav)
	// set the current navigation state
	s.nav = child
	return s.uponArrivalAtNewScreen(from)
}

// goBack closes the active screen, running its onDeparture, and returns to
// its parent with the selection the parent had. Going back from the root is a
// programming error and panics.
func (s *State) goBack() tea.Cmd {
	if len(s.navStack) == 0 {
		panic(fmt.Sprintf("goBack from screen %d with nothing open beneath it", s.nav.screen))
	}
	last := len(s.navStack) - 1
	parent := s.navStack[last]
	departure := s.uponDepartureFromScreen(parent.screen)
	s.nav = parent
	s.navStack = s.navStack[:last]
	return departure
}

// uponDepartureFromScreen does the work the active screen needs when goBack
// closes it for the screen given, its parent.
func (s *State) uponDepartureFromScreen(to config.Screen) tea.Cmd {
	if depart := s.activeScreenBehaviour().onDeparture; depart != nil {
		return depart(s, to)
	}
	return nil
}

// uponArrivalAtNewScreen starts the work the active screen needs when
// openScreen opens it from the screen given, its parent. Returning to a screen
// from its child never calls it, because the screen was never closed.
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
