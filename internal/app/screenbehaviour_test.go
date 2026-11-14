package app

import (
	"shellforge/internal/config"
	"testing"
)

// TestEveryScreenHasABehaviour fails when a screen is added to config without
// an entry in behaviourOf, which would otherwise make it behave as the main
// menu, or when a list screen is missing what list navigation needs.
func TestEveryScreenHasABehaviour(t *testing.T) {
	// DeleteSandboxesScreen is the last config.Screen; extend this when adding one.
	for screen := config.MenuScreen; screen <= config.DeleteSandboxesScreen; screen++ {
		behaviour, ok := behaviourOf(screen)
		if !ok {
			t.Errorf("screen %d has no entry in behaviourOf", screen)
			continue
		}
		if behaviour.view == nil {
			t.Errorf("screen %d has no view", screen)
		}
		if behaviour.keyPress == nil && (behaviour.itemCount == nil || behaviour.selectItem == nil) {
			t.Errorf("list screen %d needs both itemCount and selectItem", screen)
		}
	}
}
