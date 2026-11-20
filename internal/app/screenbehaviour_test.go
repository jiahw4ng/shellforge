package app

import (
	"shellforge/internal/config"
	"testing"
)

// TestEveryScreenHasABehaviour fails when a screen is added to config without
// a case in behaviourOf (its empty behaviour has no view), when a screen sets
// both or neither of onKeyPress and list, when a list screen is missing what
// list navigation needs, when a screen's help is set up wrongly, or when the
// parents do not form a tree rooted at the main menu.
//
// Help is off in helpBehaviour's zero value, so a screen that forgets it would
// silently lose the help key. Only the standalone sandbox goes without help,
// because there the key belongs to the shell.
func TestEveryScreenHasABehaviour(t *testing.T) {
	for screen := range config.TotalScreenCount {
		behaviour := behaviourOf(screen)
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

		// Only the two terminal screens show a terminal, and each must say
		// how big it is.
		wantTerm := screen == config.SandboxTerminalScreen || screen == config.LessonTerminalScreen
		if (behaviour.term != nil) != wantTerm {
			t.Errorf("screen %d has term = %t, want %t", screen, behaviour.term != nil, wantTerm)
		}
		if behaviour.term != nil && behaviour.term.size == nil {
			t.Errorf("terminal screen %d has no term.size", screen)
		}

		// Every screen but the menu has a parent, and following parents
		// reaches the menu; more steps than there are screens means a cycle.
		if (behaviour.parent == nil) != (screen == config.MainMenuScreen) {
			t.Errorf("screen %d parent = %v; only the main menu has none", screen, behaviour.parent)
			continue
		}
		ancestor, steps := screen, 0
		for ancestor != config.MainMenuScreen && steps <= int(config.TotalScreenCount) {
			parent := behaviourOf(ancestor).parent
			if parent == nil {
				break
			}
			ancestor, steps = *parent, steps+1
		}
		if ancestor != config.MainMenuScreen {
			t.Errorf("screen %d does not reach the main menu through its parents", screen)
		}
	}
}
