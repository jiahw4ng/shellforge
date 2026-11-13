package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// openManageSandboxes selects Manage sandboxes in Settings and returns the
// state and the command that would run the listing.
func openManageSandboxes(t *testing.T) (State, tea.Cmd) {
	t.Helper()
	return openManageSandboxesFrom(t, New())
}

func TestManageSandboxesStartsListingWithSpinner(t *testing.T) {
	model, command := openManageSandboxes(t)

	if model.nav.screen != config.ManageSandboxesScreen {
		t.Fatalf("screen = %d, want Manage sandboxes", model.nav.screen)
	}
	if !model.settings.sandboxes.isLoading || command == nil {
		t.Fatalf("isLoading = %v, command = %v; want a listing in progress", model.settings.sandboxes.isLoading, command)
	}
	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "Looking for Shellforge sandboxes...") {
		t.Errorf("loading view = %q, want the loading line", view)
	}
}

func TestManageSandboxesShowsListedSandboxes(t *testing.T) {
	model, _ := openManageSandboxes(t)
	listedAt := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	model = updateModel(t, model, sandboxesListedMsg{
		sandboxes: []container.ManagedSandbox{
			{Name: "shellforge-new", State: "running", CreatedAt: listedAt.Add(-12 * time.Minute)},
			{Name: "shellforge-old", State: "exited", CreatedAt: listedAt.Add(-50 * time.Hour)},
		},
		listedAt: listedAt,
		gen:      model.settings.sandboxes.gen,
	})

	if model.settings.sandboxes.isLoading {
		t.Error("still loading after the listing arrived")
	}
	view := ansi.Strip(model.activeScreenView())
	for _, want := range []string{"shellforge-new", "running", "12m ago", "shellforge-old", "exited", "2d ago", config.BackMenuLabel} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q:\n%s", want, view)
		}
	}
}

func TestManageSandboxesShowsWhenNoneExist(t *testing.T) {
	model, _ := openManageSandboxes(t)
	model = updateModel(t, model, sandboxesListedMsg{gen: model.settings.sandboxes.gen})

	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "No Shellforge sandboxes found.") {
		t.Errorf("view = %q, want the empty message", view)
	}
}

func TestManageSandboxesExplainsUnavailableDocker(t *testing.T) {
	model, _ := openManageSandboxes(t)
	err := &container.ContainerUnavailableError{Reason: container.ReasonDockerUnreachable, Cause: errors.New("cannot connect")}
	model = updateModel(t, model, sandboxesListedMsg{err: err, gen: model.settings.sandboxes.gen})

	view := ansi.Strip(model.activeScreenView())
	if !strings.Contains(view, "Docker") || !strings.Contains(view, config.BackMenuLabel) {
		t.Errorf("view = %q, want the Docker problem and Back", view)
	}
}

func TestManageSandboxesIgnoresStaleListings(t *testing.T) {
	t.Run("earlier visit", func(t *testing.T) {
		model, _ := openManageSandboxes(t)
		stale := model.settings.sandboxes.gen - 1
		model = updateModel(t, model, sandboxesListedMsg{
			sandboxes: []container.ManagedSandbox{{Name: "shellforge-stale"}},
			gen:       stale,
		})
		if !model.settings.sandboxes.isLoading || len(model.settings.sandboxes.found) != 0 {
			t.Errorf("stale listing applied: %+v", model.settings.sandboxes)
		}
	})

	t.Run("after leaving", func(t *testing.T) {
		model, _ := openManageSandboxes(t)
		gen := model.settings.sandboxes.gen
		model = updateModel(t, model, keyPress(tea.KeyEnter, "")) // Back
		model = updateModel(t, model, sandboxesListedMsg{
			sandboxes: []container.ManagedSandbox{{Name: "shellforge-late"}},
			gen:       gen,
		})
		if model.settings.sandboxes.isLoading || len(model.settings.sandboxes.found) != 0 {
			t.Errorf("listing after leaving = %+v, want stopped and empty", model.settings.sandboxes)
		}
	})
}

func TestManageSandboxesBackReturnsToSettings(t *testing.T) {
	model, _ := openManageSandboxes(t)
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	result := updated.(State)

	if result.nav.screen != config.SettingsScreen || result.nav.selection != settingsManageSandboxesIndex {
		t.Errorf("screen = %d, selection = %d; want Settings with Manage sandboxes selected", result.nav.screen, result.nav.selection)
	}
	if command != nil {
		t.Error("leaving Manage sandboxes scheduled work")
	}
}

func TestManageSandboxesArrowKeysDoNotRestartListing(t *testing.T) {
	model, _ := openManageSandboxes(t)
	gen := model.settings.sandboxes.gen
	updated, command := model.Update(keyPress(tea.KeyDown, ""))
	result := updated.(State)

	if result.settings.sandboxes.gen != gen || command != nil {
		t.Errorf("gen %d -> %d, command %v; want the same listing", gen, result.settings.sandboxes.gen, command)
	}
}

// listedManageSandboxes opens Manage sandboxes and delivers a listing of two
// sandboxes, so the items are Select all, two rows, Delete selected, Back.
func listedManageSandboxes(t *testing.T) State {
	t.Helper()
	model, _ := openManageSandboxes(t)
	return updateModel(t, model, sandboxesListedMsg{
		sandboxes: []container.ManagedSandbox{
			{ID: "aaa", Name: "shellforge-a", State: "running"},
			{ID: "bbb", Name: "shellforge-b", State: "exited"},
		},
		gen: model.settings.sandboxes.gen,
	})
}

func spacePress() tea.KeyPressMsg {
	return keyPress(tea.KeySpace, " ")
}

func TestManageSandboxesListingSelectsSelectAll(t *testing.T) {
	model := listedManageSandboxes(t)
	if model.nav.selection != 0 {
		t.Errorf("selection = %d, want 0 (Select all)", model.nav.selection)
	}
	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "▶ [ ] Select all") {
		t.Errorf("view does not select Select all:\n%s", view)
	}
}

func TestManageSandboxesSpaceTicksAndUnticksARow(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, "")) // shellforge-a
	model = updateModel(t, model, spacePress())

	if !model.settings.sandboxes.chosen["aaa"] || model.settings.sandboxes.chosen["bbb"] {
		t.Fatalf("chosen = %v, want only aaa", model.settings.sandboxes.chosen)
	}
	view := ansi.Strip(model.activeScreenView())
	for _, want := range []string{"[x] shellforge-a", "[ ] shellforge-b", "Delete selected (1)"} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q:\n%s", want, view)
		}
	}

	model = updateModel(t, model, spacePress())
	if model.settings.sandboxes.chosen["aaa"] {
		t.Error("second Space did not untick shellforge-a")
	}
}

func TestManageSandboxesEnterTicksARow(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, keyPress(tea.KeyDown, "")) // shellforge-b
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.ManageSandboxesScreen || !model.settings.sandboxes.chosen["bbb"] {
		t.Errorf("screen = %d, chosen = %v; want shellforge-b ticked on Manage sandboxes", model.nav.screen, model.settings.sandboxes.chosen)
	}
}

func TestManageSandboxesSelectAllTicksThenUnticksEveryRow(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, spacePress()) // tick shellforge-a only
	model = updateModel(t, model, keyPress(tea.KeyUp, ""))

	model = updateModel(t, model, spacePress())
	if !model.settings.sandboxes.chosen["aaa"] || !model.settings.sandboxes.chosen["bbb"] {
		t.Fatalf("after Select all, chosen = %v; want both", model.settings.sandboxes.chosen)
	}
	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "[x] Select all") {
		t.Errorf("view does not tick Select all:\n%s", view)
	}

	model = updateModel(t, model, spacePress())
	if model.settings.sandboxes.chosen["aaa"] || model.settings.sandboxes.chosen["bbb"] {
		t.Errorf("after a second Select all, chosen = %v; want none", model.settings.sandboxes.chosen)
	}
}

func TestManageSandboxesSpaceOnBackDoesNotLeave(t *testing.T) {
	model := listedManageSandboxes(t)
	model.nav.selection = 4 // Back
	model = updateModel(t, model, spacePress())

	if model.nav.screen != config.ManageSandboxesScreen {
		t.Errorf("screen = %d, want Space on Back to stay on Manage sandboxes", model.nav.screen)
	}
}

func TestManageSandboxesEnterOnBackAfterListingReturnsToSettings(t *testing.T) {
	model := listedManageSandboxes(t)
	model.nav.selection = 4 // Back
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.SettingsScreen || model.nav.selection != settingsManageSandboxesIndex {
		t.Errorf("screen = %d, selection = %d; want Settings with Manage sandboxes selected", model.nav.screen, model.nav.selection)
	}
}

func TestManageSandboxesNewListingClearsTicks(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, spacePress()) // Select all
	model.nav.selection = 4
	model = updateModel(t, model, keyPress(tea.KeyEnter, "")) // Back
	model, _ = openManageSandboxesFrom(t, model)

	if len(model.settings.sandboxes.chosen) != 0 {
		t.Errorf("chosen = %v after reopening, want none", model.settings.sandboxes.chosen)
	}
}

// openManageSandboxesFrom presses Enter on Manage sandboxes in Settings.
func openManageSandboxesFrom(t *testing.T, model State) (State, tea.Cmd) {
	t.Helper()
	model.nav = navigationState{screen: config.SettingsScreen, selection: settingsManageSandboxesIndex}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	return updated.(State), command
}

func TestManageSandboxesHelpShowsSpace(t *testing.T) {
	model := listedManageSandboxes(t)
	model.help.ShowAll = true
	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "Space") || !strings.Contains(view, "tick/untick") {
		t.Errorf("expanded help does not mention Space:\n%s", view)
	}
}
