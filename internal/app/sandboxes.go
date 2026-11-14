package app

import (
	"context"
	"fmt"
	"log/slog"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"shellforge/internal/logging"
	"shellforge/internal/screens"
	"shellforge/internal/terminal"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/samber/lo"
)

// This file holds everything for the Manage sandboxes and Delete sandboxes
// screens: their state, the Docker listing and deletion commands, key
// handling, async result handling, and rendering.

// ---------------------------------------------------------------------------
// State
//
// sandboxListState lives in settingsState.sandboxes. It holds the latest
// listing and which rows are ticked. Ticks are kept by container ID, so they
// survive the rows being re-rendered but are cleared by every new listing.
// ---------------------------------------------------------------------------

type sandboxListState struct {
	// isLoading reports that a listing command is in progress.
	isLoading bool
	// gen identifies one listing. Each visit to the screen starts a new one,
	// so a slow result from an earlier visit cannot replace a newer one.
	gen uint64
	// found is the latest listing, newest first.
	found []container.ManagedSandbox
	// err is why the latest listing failed, if it did.
	err error
	// listedAt is when the latest listing was taken; ages are measured from it.
	listedAt time.Time
	// spinner animates while isLoading.
	spinner spinner.Model
	// chosen holds the IDs of the ticked sandboxes. Every listing clears it.
	chosen map[string]bool
	// isDeleting reports that the ticked sandboxes are being removed. Only
	// Ctrl+C is handled meanwhile.
	isDeleting bool
	// message reports the latest deletion; failed marks it as an error. Every
	// listing clears it, so it must be set after the refresh begins.
	message string
	failed  bool
}

// chosenSandboxes returns the ticked sandboxes in listing order.
func (l sandboxListState) chosenSandboxes() []container.ManagedSandbox {
	return lo.Filter(l.found, func(sandbox container.ManagedSandbox, _ int) bool { return l.chosen[sandbox.ID] })
}

// toggle ticks or unticks the listed sandbox at index.
func (l *sandboxListState) toggle(index int) {
	id := l.found[index].ID
	l.chosen[id] = !l.chosen[id]
}

// toggleAll ticks every listed sandbox, or unticks them all when every one is
// already ticked.
func (l *sandboxListState) toggleAll() {
	all := lo.EveryBy(l.found, func(sandbox container.ManagedSandbox) bool { return l.chosen[sandbox.ID] })
	for _, sandbox := range l.found {
		l.chosen[sandbox.ID] = !all
	}
}

// ---------------------------------------------------------------------------
// Listing
//
// Flow: begin() bumps gen and schedules listSandboxes() -> Docker runs off the
// Update loop -> sandboxesListedMsg -> handleSandboxesListed() stores the rows
// if the message still matches the current gen and the listing was not
// stopped by leaving the screen.
// ---------------------------------------------------------------------------

// begin starts a new listing and returns the commands that run it and
// animate its spinner.
func (l *sandboxListState) begin() tea.Cmd {
	l.gen++
	l.isLoading = true
	l.found = nil
	l.err = nil
	l.chosen = make(map[string]bool)
	l.message = ""
	l.failed = false
	l.spinner = spinner.New(spinner.WithSpinner(spinner.Dot))
	return tea.Batch(listSandboxes(l.gen), l.spinner.Tick)
}

// listSandboxes lists every Shellforge sandbox without blocking the TUI. Each
// Docker step is bounded inside container.ListManaged.
func listSandboxes(generation uint64) tea.Cmd {
	return func() tea.Msg {
		sandboxes, err := container.ListManaged(context.Background())
		return sandboxesListedMsg{sandboxes: sandboxes, err: err, listedAt: time.Now(), gen: generation}
	}
}

// sandboxesListedMsg carries the result of one Manage sandboxes listing.
type sandboxesListedMsg struct {
	sandboxes []container.ManagedSandbox
	err       error
	listedAt  time.Time
	gen       uint64
}

// handleSandboxesListed records the current listing's result. Results from an
// earlier visit, or arriving after the learner left the screen (which stops
// the listing), are ignored.
func (s State) handleSandboxesListed(msg sandboxesListedMsg) (tea.Model, tea.Cmd) {
	listing := &s.settings.sandboxes
	if msg.gen != listing.gen || !listing.isLoading {
		return s, nil
	}
	listing.isLoading = false
	listing.found = msg.sandboxes
	listing.err = msg.err
	listing.listedAt = msg.listedAt
	// The items change once the rows arrive, so start again from the top.
	s.nav.selection = 0
	if msg.err != nil {
		slog.Error("Could not list Shellforge sandboxes", logErrorKey, msg.err)
	}
	return s, nil
}

// ---------------------------------------------------------------------------
// Key handling
//
// Only Enter is screen-specific; every other key, and Enter on Back, falls
// through to the shared handleNavigationKeyPress. Item positions come from
// screens.ManageSandboxes.Item, so rendering and key handling share one layout.
// ---------------------------------------------------------------------------

// handleManageSandboxesKeyPress acts on Enter: it ticks or unticks a row or
// every row, opens the deletion confirmation, or returns to Settings from Back.
// Other keys navigate.
func (s State) handleManageSandboxesKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// if the user did not press Enter, handle navigation normally
	if !key.Matches(msg, config.Keys.SelectItem) {
		return s.handleNavigationKeyPress(msg)
	}
	screen := s.manageSandboxesScreen()
	kind, row := screen.Item(s.nav.selection)
	switch kind {
	case screens.SandboxSelectAllItem:
		// [ ] Select all
		s.settings.sandboxes.toggleAll()
	case screens.SandboxRowItem:
		// [ ] shellforge-wxyz running x ago
		s.settings.sandboxes.toggle(row)
	case screens.SandboxDeleteItem:
		// Delete selected (x)
		if screen.ChosenCount() > 0 {
			return s, s.navigate(config.DeleteSandboxesScreen, 0)
		}
		// Do nothing if no sandboxes are ticked
	case screens.SandboxBackItem:
		// Back
		return s.handleNavigationKeyPress(msg)
	}
	return s, nil
}

// handleDeleteSandboxesKeyPress navigates the deletion confirmation. Enter on
// "No, go back" returns to Manage sandboxes with the ticks kept; Enter on the
// confirming choice deletes the ticked sandboxes.
func (s State) handleDeleteSandboxesKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if !key.Matches(msg, config.Keys.SelectItem) {
		return s.handleNavigationKeyPress(msg)
	}
	if s.nav.selection == 1 {
		return s, s.settings.sandboxes.beginDeleting()
	}
	back := s.navigate(config.ManageSandboxesScreen, len(s.settings.sandboxes.found)+1) // Delete selected
	return s, back
}

// ---------------------------------------------------------------------------
// Deletion
//
// Flow: confirming calls beginDeleting, which sets isDeleting (handleKeyPress
// then ignores everything but Ctrl+C) and schedules deleteSandboxes ->
// sandboxesDeletedMsg -> handleSandboxesDeleted returns to Manage sandboxes,
// starts a fresh listing, and reports the outcome. The report is written
// after begin because begin clears the message.
// ---------------------------------------------------------------------------

// beginDeleting removes the ticked sandboxes and animates the spinner meanwhile.
func (l *sandboxListState) beginDeleting() tea.Cmd {
	l.isDeleting = true
	ids := lo.Map(l.chosenSandboxes(), func(sandbox container.ManagedSandbox, _ int) string { return sandbox.ID })
	return tea.Batch(deleteSandboxes(ids), l.spinner.Tick)
}

// deleteSandboxes removes the ticked sandboxes without blocking the TUI. Each
// Docker step is bounded inside container.RemoveManaged.
func deleteSandboxes(ids []string) tea.Cmd {
	return func() tea.Msg {
		failed, err := container.RemoveManaged(context.Background(), ids)
		return sandboxesDeletedMsg{failed: failed, err: err}
	}
}

// sandboxesDeletedMsg reports a deletion of ticked sandboxes: the IDs Docker
// could not remove, or why nothing was removed.
type sandboxesDeletedMsg struct {
	failed []string
	err    error
}

// handleSandboxesDeleted returns to Manage sandboxes, reports which sandboxes
// were deleted or could not be, and lists them again. When Docker could not
// be used at all, the new listing shows the Docker problem and its remedy.
func (s State) handleSandboxesDeleted(msg sandboxesDeletedMsg) (tea.Model, tea.Cmd) {
	listing := &s.settings.sandboxes
	if !listing.isDeleting {
		return s, nil
	}
	listing.isDeleting = false
	chosen := listing.chosenSandboxes()
	names := lo.SliceToMap(listing.found, func(sandbox container.ManagedSandbox) (string, string) { return sandbox.ID, sandbox.Name })

	// Coming back from the confirmation does not start a listing (see arrive),
	// so start it here: the outcome below must be written after begin clears it.
	refresh := tea.Batch(s.navigate(config.ManageSandboxesScreen, 0), listing.begin())
	switch {
	case msg.err != nil:
		slog.Error("Could not delete Shellforge sandboxes", logErrorKey, msg.err)
		listing.message = noSandboxesDeletedMessage
		listing.failed = true
	case len(msg.failed) > 0:
		failedNames := lo.Map(msg.failed, func(id string, _ int) string { return lo.ValueOr(names, id, id) })
		listing.message = fmt.Sprintf(sandboxesNotDeletedFormat, strings.Join(failedNames, ", "))
		listing.failed = true
	default:
		listing.message = deletedSandboxesMessage(len(chosen))
	}
	return s, refresh
}

const (
	// noSandboxesDeletedMessage reports a deletion that could not use Docker.
	noSandboxesDeletedMessage = "No sandboxes were deleted."
	// sandboxesNotDeletedFormat names the sandboxes Docker could not remove.
	sandboxesNotDeletedFormat = "Could not delete: %s"
)

// deletedSandboxesMessage reports how many sandboxes were deleted.
func deletedSandboxesMessage(count int) string {
	if count == 1 {
		return "Deleted 1 sandbox."
	}
	return fmt.Sprintf("Deleted %d sandboxes.", count)
}

// ---------------------------------------------------------------------------
// View
//
// Translates sandboxListState into the presentation structs in
// internal/screens. No I/O happens here; the screens package does the drawing.
// ---------------------------------------------------------------------------

// manageSandboxesView shows the latest sandbox listing, its spinner while it
// runs, or why it failed.
func (s State) manageSandboxesView(helpWidth int) string {
	screen := s.manageSandboxesScreen()
	screen.Help = s.navigationHelp(helpWidth)
	return screen.View()
}

// manageSandboxesScreen describes Manage sandboxes without its help line. Key
// handling uses it too, so items are laid out in one place.
func (s State) manageSandboxesScreen() screens.ManageSandboxes {
	listing := s.settings.sandboxes
	return screens.ManageSandboxes{
		Sandboxes: listing.found,
		ListedAt:  listing.listedAt,
		Loading:   listing.isLoading,
		Spinner:   listing.spinner.View(),
		Problem:   sandboxListingProblem(listing.err),
		Chosen:    listing.chosen,
		Selection: s.nav.selection,
		Message:   listing.message,
		Failed:    listing.failed,
	}
}

// deleteSandboxesView asks the learner to confirm deleting the ticked
// sandboxes, hiding the help line while they are being deleted.
func (s State) deleteSandboxesView(helpWidth int) string {
	listing := s.settings.sandboxes
	screen := screens.DeleteSandboxesConfirmation{
		Sandboxes: listing.chosenSandboxes(),
		Selection: s.nav.selection,
		Deleting:  listing.isDeleting,
		Spinner:   listing.spinner.View(),
	}
	if !listing.isDeleting {
		screen.Help = s.navigationHelp(helpWidth)
	}
	return screen.View()
}

// sandboxListingProblem explains a failed listing: a recognised Docker problem
// gets its summary and remedy, anything else shows its own text.
func sandboxListingProblem(err error) string {
	if err == nil {
		return ""
	}
	if message, ok := terminal.Describe(err); ok {
		return screens.TerminalProblem{
			Summary: message.Summary,
			Remedy:  message.Remedy,
			Detail:  message.Detail,
			LogPath: logging.Path(),
		}.View()
	}
	return screens.TerminalFailure("", err)
}
