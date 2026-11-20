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

// openChildScreen, goBackToParentScreen, and goBackWithReport are the only
// ways to change screen. Screens form a tree (see screenBehaviour.parent) and
// navStack is the path from the root to the active screen, so a screen is open
// from openChildScreen until it goes back, and its parent stays open
// underneath meanwhile.

// openChildScreen opens childScreen, a child of the active screen, on its
// first item, and returns the work its onArrival starts; onArrival may select
// another item. The active screen stays open underneath with its selection,
// so going back returns to it as it was. Opening a screen that is not a child
// of the active one is a programming error and panics.
func (s *State) openChildScreen(childScreen config.Screen) tea.Cmd {
	curr := s.nav.screen
	if parent := behaviourOf(childScreen).parent; parent == nil || *parent != curr {
		panic(fmt.Sprintf("screen %d opened from screen %d, which is not its parent", childScreen, curr))
	}
	// push the parent into the existing stack
	s.navStack = append(slices.Clip(s.navStack), s.nav)
	// set the current navigation state
	s.nav = navigationState{screen: childScreen, selection: 0}
	return s.uponArrivalAtNewScreen()
}

// goBackToParentScreen closes the active screen, running its onDeparture, and
// returns to its parent with the selection the parent had, running the
// parent's onReturnFromChild with no report. Going back from the root is a
// programming error and panics.
func (s *State) goBackToParentScreen() tea.Cmd {
	return s.goBackWithReport(nil)
}

// goBackWithReport is goBackToParentScreen, handing report to the parent's
// onReturnFromChild so the parent can show the outcome of the child's work and
// react to it.
func (s *State) goBackWithReport(report *childReport) tea.Cmd {
	if len(s.navStack) == 0 {
		panic(fmt.Sprintf("going back from screen %d with nothing open beneath it", s.nav.screen))
	}
	topIdx := len(s.navStack) - 1
	parent := s.navStack[topIdx]
	departure := s.uponDepartureFromScreen()
	s.nav = parent
	s.navStack = s.navStack[:topIdx]
	return tea.Batch(departure, s.uponReturnFromChild(report))
}

// uponReturnFromChild runs the active (parent) screen's onReturnFromChild
// after going back closed its child, passing on the child's report.
func (s *State) uponReturnFromChild(report *childReport) tea.Cmd {
	if onReturnFromChild := s.activeScreenBehaviour().onReturnFromChild; onReturnFromChild != nil {
		return onReturnFromChild(s, report)
	}
	return nil
}

// uponDepartureFromScreen does the work the active screen needs when going
// back closes it.
func (s *State) uponDepartureFromScreen() tea.Cmd {
	if depart := s.activeScreenBehaviour().onDeparture; depart != nil {
		return depart(s)
	}
	return nil
}

// uponArrivalAtNewScreen starts the work the active screen needs when
// openChildScreen opens it. Returning to a screen from its child never calls
// it, because the screen was never closed.
func (s *State) uponArrivalAtNewScreen() tea.Cmd {
	if arrive := s.activeScreenBehaviour().onArrival; arrive != nil {
		return arrive(s)
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
