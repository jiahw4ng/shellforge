package app

import (
	"shellforge/internal/config"
	"shellforge/internal/container"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// showing puts model on screen with selection selected, as if every screen
// above it in the tree had been opened in turn from the main menu, each with
// its first item selected. Tests that check where goBack returns set the
// parents' selections in model.navStack themselves.
func showing(model State, screen config.Screen, selection int) State {
	var path []navigationState
	for parent := behaviourOf(screen).parent; parent != nil; parent = behaviourOf(*parent).parent {
		path = append(path, navigationState{screen: *parent})
	}
	slices.Reverse(path)
	model.navStack = path
	model.nav = navigationState{screen: screen, selection: selection}
	return model
}

// openedFrom is showing, with the screen opened from its parent while the
// parent had parentSelection selected, so goBack returns to that item.
func openedFrom(model State, screen config.Screen, selection, parentSelection int) State {
	model = showing(model, screen, selection)
	model.navStack[len(model.navStack)-1].selection = parentSelection
	return model
}

// TestBackRestoresTheSelectionAtEveryLevel walks the deepest path, Main menu
// -> Settings -> Manage sandboxes -> Delete sandboxes, with real key presses,
// then back out again. Each Back must land on the item that opened the screen,
// and the stack must be empty again on the menu.
func TestBackRestoresTheSelectionAtEveryLevel(t *testing.T) {
	enter := keyPress(tea.KeyEnter, "")
	choose := func(model State, item int) State {
		t.Helper()
		model.nav.selection = item
		return updateModel(t, model, enter)
	}
	want := func(model State, screen config.Screen, selection int, step string) {
		t.Helper()
		if model.nav.screen != screen || model.nav.selection != selection {
			t.Fatalf("%s: screen %d item %d, want screen %d item %d", step, model.nav.screen, model.nav.selection, screen, selection)
		}
	}

	model := choose(New(), menuSettingsIndex)
	model = choose(model, settingsManageSandboxesIndex)
	model = updateModel(t, model, sandboxesListedMsg{
		sandboxes: []container.ManagedSandbox{{ID: "aaa", Name: "shellforge-a"}},
		gen:       model.sandboxes.gen,
	})
	model = choose(model, 0) // Select all
	deleteItem := model.manageSandboxesScreen().DeleteItem()
	model = choose(model, deleteItem)
	want(model, config.DeleteSandboxesScreen, 0, "open the deletion confirmation")
	if len(model.navStack) != 3 {
		t.Fatalf("stack depth = %d on the deletion confirmation, want 3", len(model.navStack))
	}

	model = choose(model, 0) // No, go back
	want(model, config.ManageSandboxesScreen, deleteItem, "back from the confirmation")
	model = choose(model, model.manageSandboxesScreen().ItemCount()-1) // Back
	want(model, config.SettingsScreen, settingsManageSandboxesIndex, "back from Manage sandboxes")
	model = choose(model, len(config.SettingsItems)-1) // Back
	want(model, config.MenuScreen, menuSettingsIndex, "back from Settings")
	if len(model.navStack) != 0 {
		t.Fatalf("stack = %v on the main menu, want empty", model.navStack)
	}
}

// TestOpenScreenPanicsForAScreenThatIsNotAChild opens Theme from the main
// menu, which skips Settings; the tree only allows Settings to open it.
func TestOpenScreenPanicsForAScreenThatIsNotAChild(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("opening Theme from the main menu did not panic")
		}
	}()
	model := New()
	model.openScreen(navigationState{screen: config.ThemeScreen})
}

// TestGoBackPanicsOnTheMainMenu: the menu is the root, so there is nothing to
// go back to.
func TestGoBackPanicsOnTheMainMenu(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("goBack on the main menu did not panic")
		}
	}()
	model := New()
	model.goBack()
}

// TestOpenScreenDoesNotShareTheStackWithEarlierStates opens screens from two
// copies of one State. Going back first leaves the stack with spare capacity,
// which a plain append would fill in place, so both copies would write the same
// slot and the second push would change the first copy's saved selection.
func TestOpenScreenDoesNotShareTheStackWithEarlierStates(t *testing.T) {
	settings := showing(New(), config.ThemeScreen, 0)
	settings.goBack() // the stack is now [menu] with room for one more

	theme, cursor := settings, settings
	theme.nav.selection = settingsThemeIndex
	theme.openScreen(navigationState{screen: config.ThemeScreen})
	cursor.nav.selection = settingsCursorStyleIndex
	cursor.openScreen(navigationState{screen: config.CursorStyleScreen})

	theme.goBack()
	if theme.nav.selection != settingsThemeIndex {
		t.Fatalf("back from Theme on Settings item %d, want %d: the copies share a stack", theme.nav.selection, settingsThemeIndex)
	}
}
