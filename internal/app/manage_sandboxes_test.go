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
	if !model.sandboxes.isLoading || command == nil {
		t.Fatalf("isLoading = %v, command = %v; want a listing in progress", model.sandboxes.isLoading, command)
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
		gen:      model.sandboxes.gen,
	})

	if model.sandboxes.isLoading {
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
	model = updateModel(t, model, sandboxesListedMsg{gen: model.sandboxes.gen})

	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "No Shellforge sandboxes found.") {
		t.Errorf("view = %q, want the empty message", view)
	}
}

func TestManageSandboxesExplainsUnavailableDocker(t *testing.T) {
	model, _ := openManageSandboxes(t)
	err := &container.ContainerUnavailableError{Reason: container.ReasonDockerUnreachable, Cause: errors.New("cannot connect")}
	model = updateModel(t, model, sandboxesListedMsg{err: err, gen: model.sandboxes.gen})

	view := ansi.Strip(model.activeScreenView())
	if !strings.Contains(view, "Docker") || !strings.Contains(view, config.BackMenuLabel) {
		t.Errorf("view = %q, want the Docker problem and Back", view)
	}
}

func TestManageSandboxesIgnoresStaleListings(t *testing.T) {
	t.Run("earlier visit", func(t *testing.T) {
		model, _ := openManageSandboxes(t)
		stale := model.sandboxes.gen - 1
		model = updateModel(t, model, sandboxesListedMsg{
			sandboxes: []container.ManagedSandbox{{Name: "shellforge-stale"}},
			gen:       stale,
		})
		if !model.sandboxes.isLoading || len(model.sandboxes.found) != 0 {
			t.Errorf("stale listing applied: %+v", model.sandboxes)
		}
	})

	t.Run("after leaving", func(t *testing.T) {
		model, _ := openManageSandboxes(t)
		gen := model.sandboxes.gen
		model = updateModel(t, model, keyPress(tea.KeyEnter, "")) // Back
		model = updateModel(t, model, sandboxesListedMsg{
			sandboxes: []container.ManagedSandbox{{Name: "shellforge-late"}},
			gen:       gen,
		})
		if model.sandboxes.isLoading || len(model.sandboxes.found) != 0 {
			t.Errorf("listing after leaving = %+v, want stopped and empty", model.sandboxes)
		}
	})
}

// TestManageSandboxesIgnoresListingFromBeforeReopeningSettings covers a late
// listing from an earlier visit when the learner went back to the main menu
// and opened Settings again in between, which once restarted the generations.
func TestManageSandboxesIgnoresListingFromBeforeReopeningSettings(t *testing.T) {
	model, _ := openManageSandboxes(t)
	earlier := model.sandboxes.gen
	model = updateModel(t, model, enterPress()) // Back to Settings
	model.nav = navigationState{screen: config.MainMenuScreen, selection: menuSettingsIndex}
	model = updateModel(t, model, enterPress()) // open Settings from the menu
	model, _ = openManageSandboxesFrom(t, model)

	model = updateModel(t, model, sandboxesListedMsg{
		sandboxes: []container.ManagedSandbox{{Name: "shellforge-earlier"}},
		gen:       earlier,
	})
	if !model.sandboxes.isLoading || len(model.sandboxes.found) != 0 {
		t.Errorf("listing from the earlier visit applied: %+v", model.sandboxes)
	}
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
	gen := model.sandboxes.gen
	updated, command := model.Update(keyPress(tea.KeyDown, ""))
	result := updated.(State)

	if result.sandboxes.gen != gen || command != nil {
		t.Errorf("gen %d -> %d, command %v; want the same listing", gen, result.sandboxes.gen, command)
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
		gen: model.sandboxes.gen,
	})
}

func enterPress() tea.KeyPressMsg {
	return keyPress(tea.KeyEnter, "")
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

func TestManageSandboxesEnterTicksAndUnticksARow(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, "")) // shellforge-a
	model = updateModel(t, model, enterPress())

	if !model.sandboxes.chosen["aaa"] || model.sandboxes.chosen["bbb"] {
		t.Fatalf("chosen = %v, want only aaa", model.sandboxes.chosen)
	}
	view := ansi.Strip(model.activeScreenView())
	for _, want := range []string{"[x] shellforge-a", "[ ] shellforge-b", "Delete selected (1)"} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q:\n%s", want, view)
		}
	}

	model = updateModel(t, model, enterPress())
	if model.sandboxes.chosen["aaa"] {
		t.Error("second Enter did not untick shellforge-a")
	}
}

func TestManageSandboxesSelectAllTicksThenUnticksEveryRow(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	model = updateModel(t, model, enterPress()) // tick shellforge-a only
	model = updateModel(t, model, keyPress(tea.KeyUp, ""))

	model = updateModel(t, model, enterPress())
	if !model.sandboxes.chosen["aaa"] || !model.sandboxes.chosen["bbb"] {
		t.Fatalf("after Select all, chosen = %v; want both", model.sandboxes.chosen)
	}
	if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, "[x] Select all") {
		t.Errorf("view does not tick Select all:\n%s", view)
	}

	model = updateModel(t, model, enterPress())
	if model.sandboxes.chosen["aaa"] || model.sandboxes.chosen["bbb"] {
		t.Errorf("after a second Select all, chosen = %v; want none", model.sandboxes.chosen)
	}
}

// TestManageSandboxesSpaceDoesNotTick checks that Enter is the only key that
// ticks a row.
func TestManageSandboxesSpaceDoesNotTick(t *testing.T) {
	model := listedManageSandboxes(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, "")) // shellforge-a
	model = updateModel(t, model, keyPress(tea.KeySpace, " "))

	if model.sandboxes.chosen["aaa"] {
		t.Error("Space ticked shellforge-a, want only Enter to tick")
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
	model = updateModel(t, model, enterPress()) // Select all
	model.nav.selection = 4
	model = updateModel(t, model, keyPress(tea.KeyEnter, "")) // Back
	model, _ = openManageSandboxesFrom(t, model)

	if len(model.sandboxes.chosen) != 0 {
		t.Errorf("chosen = %v after reopening, want none", model.sandboxes.chosen)
	}
}

// openManageSandboxesFrom presses Enter on Manage sandboxes in Settings.
func openManageSandboxesFrom(t *testing.T, model State) (State, tea.Cmd) {
	t.Helper()
	model = showing(model, config.SettingsScreen, settingsManageSandboxesIndex)
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	return updated.(State), command
}

// confirmingDeletion ticks every listed sandbox and opens the deletion
// confirmation from Delete selected.
func confirmingDeletion(t *testing.T) State {
	t.Helper()
	model := listedManageSandboxes(t)
	model = updateModel(t, model, enterPress()) // Select all
	model.nav.selection = 3                     // Delete selected
	return updateModel(t, model, keyPress(tea.KeyEnter, ""))
}

func TestManageSandboxesDeleteNeedsATick(t *testing.T) {
	model := listedManageSandboxes(t)
	model.nav.selection = 3 // Delete selected
	model = updateModel(t, model, keyPress(tea.KeyEnter, ""))

	if model.nav.screen != config.ManageSandboxesScreen {
		t.Errorf("screen = %d, want Delete selected (0) to stay on Manage sandboxes", model.nav.screen)
	}
}

func TestManageSandboxesDeleteAsksForConfirmation(t *testing.T) {
	model := confirmingDeletion(t)

	if model.nav.screen != config.DeleteSandboxesScreen || model.nav.selection != 0 {
		t.Fatalf("screen = %d, selection = %d; want the confirmation with No selected", model.nav.screen, model.nav.selection)
	}
	view := ansi.Strip(model.activeScreenView())
	for _, want := range []string{
		"Delete 2 sandboxes?",
		"shellforge-a  running: may be in use by another Shellforge window",
		"This cannot be undone.",
		"▶ No, go back",
		"Yes, delete 2 sandboxes",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not contain %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "shellforge-b  running") {
		t.Errorf("the stopped sandbox is marked as running:\n%s", view)
	}
}

func TestDeleteSandboxesGoBackKeepsTicks(t *testing.T) {
	model := confirmingDeletion(t)
	updated, command := model.Update(keyPress(tea.KeyEnter, "")) // No, go back
	model = updated.(State)

	if model.nav.screen != config.ManageSandboxesScreen || model.nav.selection != 3 || command != nil {
		t.Errorf("screen = %d, selection = %d, command = %v; want Manage sandboxes on Delete selected, no work", model.nav.screen, model.nav.selection, command)
	}
	if !model.sandboxes.chosen["aaa"] || !model.sandboxes.chosen["bbb"] {
		t.Errorf("chosen = %v, want both kept", model.sandboxes.chosen)
	}
}

func TestDeleteSandboxesConfirmStartsDeletion(t *testing.T) {
	model := confirmingDeletion(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	model = updated.(State)

	if !model.sandboxes.isDeleting || command == nil {
		t.Fatalf("isDeleting = %v, command = %v; want a deletion in progress", model.sandboxes.isDeleting, command)
	}
	view := ansi.Strip(model.activeScreenView())
	if !strings.Contains(view, "Deleting sandboxes...") {
		t.Errorf("view does not show the deletion in progress:\n%s", view)
	}

	if blocked := updateModel(t, model, keyPress(tea.KeyUp, "")); blocked.nav.selection != 1 {
		t.Errorf("selection moved to %d during deletion", blocked.nav.selection)
	}
	if _, quit := model.Update(keyPress('c', "", tea.ModCtrl)); quit == nil {
		t.Error("Ctrl+C does not quit during deletion")
	}
}

// deletingSandboxes confirms deleting both listed sandboxes.
func deletingSandboxes(t *testing.T) State {
	t.Helper()
	model := confirmingDeletion(t)
	model = updateModel(t, model, keyPress(tea.KeyDown, ""))
	return updateModel(t, model, keyPress(tea.KeyEnter, ""))
}

func TestSandboxesDeletedReportsAndRefreshes(t *testing.T) {
	tests := []struct {
		name       string
		msg        sandboxesDeletedMsg
		wantText   string
		wantFailed bool
	}{
		{name: "all deleted", msg: sandboxesDeletedMsg{}, wantText: "Deleted 2 sandboxes."},
		{name: "some failed", msg: sandboxesDeletedMsg{failed: []string{"bbb", "zzz"}}, wantText: "Could not delete: shellforge-b, zzz", wantFailed: true},
		{
			name:       "docker unavailable",
			msg:        sandboxesDeletedMsg{err: &container.ContainerUnavailableError{Reason: container.ReasonDockerUnreachable, Cause: errors.New("down")}},
			wantText:   "No sandboxes were deleted.",
			wantFailed: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := deletingSandboxes(t)
			gen := model.sandboxes.gen
			updated, command := model.Update(test.msg)
			model = updated.(State)

			listing := model.sandboxes
			if model.nav.screen != config.ManageSandboxesScreen || listing.isDeleting {
				t.Fatalf("screen = %d, isDeleting = %v; want back on Manage sandboxes", model.nav.screen, listing.isDeleting)
			}
			if !listing.isLoading || listing.gen == gen || command == nil {
				t.Errorf("isLoading = %v, gen %d -> %d, command = %v; want a fresh listing", listing.isLoading, gen, listing.gen, command)
			}
			if listing.message != test.wantText || listing.failed != test.wantFailed {
				t.Errorf("message = %q (failed %v), want %q (failed %v)", listing.message, listing.failed, test.wantText, test.wantFailed)
			}

			model = updateModel(t, model, sandboxesListedMsg{gen: listing.gen})
			if view := ansi.Strip(model.activeScreenView()); !strings.Contains(view, test.wantText) {
				t.Errorf("view after refresh does not report %q:\n%s", test.wantText, view)
			}
		})
	}
}

func TestSandboxesDeletedIgnoredWhenNotDeleting(t *testing.T) {
	model := listedManageSandboxes(t)
	updated, command := model.Update(sandboxesDeletedMsg{})
	result := updated.(State)

	if result.sandboxes.message != "" || command != nil {
		t.Errorf("message = %q, command = %v; want the unexpected result ignored", result.sandboxes.message, command)
	}
}

func TestManageSandboxesReopeningClearsDeletionMessage(t *testing.T) {
	model := deletingSandboxes(t)
	model = updateModel(t, model, sandboxesDeletedMsg{})
	model, _ = openManageSandboxesFrom(t, model)

	if model.sandboxes.message != "" {
		t.Errorf("message = %q after reopening, want none", model.sandboxes.message)
	}
}

// TestLeavingManageSandboxesAnyWayStopsTheListing calls goBackToParentScreen directly rather
// than pressing Back, so it checks that stopping the listing belongs to closing
// the screen (onDeparture), not to the Back item.
func TestLeavingManageSandboxesAnyWayStopsTheListing(t *testing.T) {
	model, _ := openManageSandboxes(t)
	gen := model.sandboxes.gen

	model.goBackToParentScreen()
	if model.sandboxes.isLoading {
		t.Fatal("listing still loading after leaving Manage sandboxes")
	}

	model = updateModel(t, model, sandboxesListedMsg{
		sandboxes: []container.ManagedSandbox{{Name: "shellforge-late"}},
		gen:       gen,
	})
	if len(model.sandboxes.found) != 0 {
		t.Errorf("late listing applied after leaving: %+v", model.sandboxes.found)
	}
}

// TestVisitingDeleteSandboxesKeepsTheListing checks that a parent stays open
// under its child: opening the deletion confirmation does not close Manage
// sandboxes (no onDeparture, so the listing keeps loading), and coming back
// does not reopen it (no onArrival, so no new listing starts).
func TestVisitingDeleteSandboxesKeepsTheListing(t *testing.T) {
	model, _ := openManageSandboxes(t)
	gen := model.sandboxes.gen

	if command := model.openChildScreen(config.DeleteSandboxesScreen); command != nil {
		t.Error("opening the confirmation scheduled work")
	}
	if !model.sandboxes.isLoading {
		t.Error("opening the confirmation stopped Manage sandboxes' listing")
	}

	if command := model.goBackToParentScreen(); command != nil {
		t.Error("returning from the confirmation scheduled work")
	}
	if model.sandboxes.gen != gen {
		t.Errorf("listing gen = %d, want %d: returning started a new listing", model.sandboxes.gen, gen)
	}
}

// TestSandboxesDeletedSelectsTheTop checks that Manage sandboxes' onReturnFromChild
// runs after going back restores its selection: Delete selected opened the
// confirmation, but the rows are about to change, so the refreshed listing
// starts from the top.
func TestSandboxesDeletedSelectsTheTop(t *testing.T) {
	model := deletingSandboxes(t)
	if restored := model.navStack[len(model.navStack)-1].selection; restored == 0 {
		t.Fatal("test setup: Manage sandboxes would already return to the top")
	}

	model = updateModel(t, model, sandboxesDeletedMsg{})
	if model.nav.screen != config.ManageSandboxesScreen || model.nav.selection != 0 {
		t.Errorf("screen = %d, selection = %d; want Manage sandboxes at the top", model.nav.screen, model.nav.selection)
	}
}
