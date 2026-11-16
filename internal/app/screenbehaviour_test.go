package app

import (
	"shellforge/internal/config"
	"testing"
)

// TestEveryScreenHasABehaviour fails when a screen is added to config without
// a case in State.behaviour (its empty behaviour has no view), when a screen
// sets both or neither of onKeyPress and list, when a list screen is missing
// what list navigation needs, or when a screen's help is set up wrongly.
//
// Help is off in helpBehaviour's zero value, so a screen that forgets it would
// silently lose the help key. Only the standalone sandbox goes without help,
// because there the key belongs to the shell.
func TestEveryScreenHasABehaviour(t *testing.T) {
	model := New()
	for screen := config.MenuScreen; screen < config.TotalScreenCount; screen++ {
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

		wantHelp := screen != config.SandboxTerminalScreen
		if behaviour.help.hasHelp != wantHelp {
			t.Errorf("screen %d help.hasHelp = %t, want %t", screen, behaviour.help.hasHelp, wantHelp)
		}
		if behaviour.help.onHelpToggle != nil && !behaviour.help.hasHelp {
			t.Errorf("screen %d sets help.onHelpToggle without help.hasHelp, so it never runs", screen)
		}
	}
}
