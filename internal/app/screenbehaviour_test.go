package app

import (
	"shellforge/internal/config"
	"testing"
)

// TestEveryScreenHasABehaviour fails when a screen is added to config without
// a case in State.behaviour (its empty behaviour has no view), when a screen
// sets both or neither of onKeyPress and list, or when a list screen is
// missing what list navigation needs.
func TestEveryScreenHasABehaviour(t *testing.T) {
	model := New()
	// DeleteSandboxesScreen is the last config.Screen; extend this when adding one.
	for screen := config.MenuScreen; screen <= config.DeleteSandboxesScreen; screen++ {
		model.nav.screen = screen
		behaviour := model.behaviour()
		if behaviour.view == nil {
			t.Errorf("screen %d has no view. It may have empty behaviour", screen)
		}
		if (behaviour.onKeyPress == nil) == (behaviour.list == nil) {
			t.Errorf("screen %d must set exactly one of onKeyPress and list", screen)
		}
		if behaviour.list != nil && (behaviour.list.itemCount == nil || behaviour.list.onSelectItem == nil) {
			t.Errorf("list screen %d needs both itemCount and onSelectItem", screen)
		}
	}
}
