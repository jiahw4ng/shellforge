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
	model := New()
	model.nav = navigationState{screen: config.SettingsScreen, selection: settingsManageSandboxesIndex}
	updated, command := model.Update(keyPress(tea.KeyEnter, ""))
	return updated.(State), command
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
