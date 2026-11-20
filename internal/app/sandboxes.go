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
// sandboxListState lives in State.sandboxes, outside settingsState, so opening
// Settings never resets its gen. It holds the latest listing and which rows
// are ticked. Ticks are kept by container ID, so they
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
	listing := &s.sandboxes
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

// handleSandboxesDeleted closes the deletion confirmation, reporting to
// Manage sandboxes which sandboxes were deleted or could not be. Manage
// sandboxes then lists them again (see its onReturnFromChild).
func (s State) handleSandboxesDeleted(msg sandboxesDeletedMsg) (tea.Model, tea.Cmd) {
	if !s.sandboxes.isDeleting {
		return s, nil
	}
	s.sandboxes.isDeleting = false
	return s, s.goBackWithReport(deletionReport(msg, s.sandboxes))
}

// deletionReport describes a finished deletion of the ticked sandboxes in
// listing, naming any that Docker could not remove. It must be built before
// a new listing replaces the rows and ticks it reads.
func deletionReport(msg sandboxesDeletedMsg, listing sandboxListState) *childReport {
	switch {
	case msg.err != nil:
		slog.Error("Could not delete Shellforge sandboxes", logErrorKey, msg.err)
		return &childReport{message: noSandboxesDeletedMessage, failed: true}
	case len(msg.failed) > 0:
		names := lo.SliceToMap(listing.found, func(sandbox container.ManagedSandbox) (string, string) { return sandbox.ID, sandbox.Name })
		failedNames := lo.Map(msg.failed, func(id string, _ int) string { return lo.ValueOr(names, id, id) })
		return &childReport{message: fmt.Sprintf(sandboxesNotDeletedFormat, strings.Join(failedNames, ", ")), failed: true}
	default:
		return &childReport{message: deletedSandboxesMessage(len(listing.chosenSandboxes()))}
	}
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

// manageSandboxesScreen describes Manage sandboxes without its help line. Key
// handling uses it too, so items are laid out in one place.
func (s State) manageSandboxesScreen() screens.ManageSandboxes {
	listing := s.sandboxes
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

// manageSandboxesBehaviour is the Manage sandboxes screen's behaviour: Select
// all, one row per sandbox, Delete selected, and Back, in the layout
// screens.ManageSandboxes.Item defines for both rendering and Enter.
func manageSandboxesBehaviour() screenBehaviour {
	return screenBehaviour{
		parent: new(config.SettingsScreen),
		list: &listBehaviour{
			// The count comes from the same layout as rendering, so it changes
			// when a listing arrives.
			itemCount: func(s State) int { return s.manageSandboxesScreen().ItemCount() },
			// Enter ticks or unticks a row or every row, opens the deletion
			// confirmation when something is ticked, or returns to Settings
			// from Back.
			onSelectItem: func(s *State) tea.Cmd {
				screen := s.manageSandboxesScreen()
				kind, row := screen.Item(s.nav.selection)
				switch kind {
				case screens.SandboxSelectAllItem:
					// [ ] Select all
					s.sandboxes.toggleAll()
				case screens.SandboxRowItem:
					// [ ] shellforge-wxyz running x ago
					s.sandboxes.toggle(row)
				case screens.SandboxDeleteItem:
					// Delete selected (x); does nothing while no sandbox is ticked
					if screen.ChosenCount() > 0 {
						return s.openChildScreen(config.DeleteSandboxesScreen)
					}
				case screens.SandboxBackItem:
					return s.goBackToParentScreen()
				}
				return nil
			},
		},
		// Closing the screen stops the listing: its spinner halts and a late
		// result is ignored. Opening the deletion confirmation leaves it open.
		onDeparture: func(s *State) tea.Cmd {
			s.sandboxes.isLoading = false
			return nil
		},
		// Opening the screen starts a new listing. Returning from the
		// deletion confirmation does not reopen it, so the listing and its
		// ticks are kept.
		onArrival: func(s *State) tea.Cmd {
			return s.sandboxes.begin()
		},
		// A finished deletion reports back here: the rows have changed, so
		// list them again from the top and show the outcome, which must be
		// written after begin clears the last one. "No, go back" reports
		// nothing, so the listing, ticks and selection stay as they were.
		onReturnFromChild: func(s *State, report *childReport) tea.Cmd {
			if report == nil {
				return nil
			}
			refresh := s.sandboxes.begin()
			s.nav.selection = 0
			s.sandboxes.message = report.message
			s.sandboxes.failed = report.failed
			return refresh
		},
		// The view shows the listing, its spinner, or the Docker problem.
		view: func(s State, helpWidth int) string {
			screen := s.manageSandboxesScreen()
			screen.Help = s.navigationHelp(helpWidth)
			return screen.View()
		},
		help: helpBehaviour{
			hasHelp: true,
		},
	}
}

// deleteSandboxesBehaviour is the behaviour of the confirmation that deletes
// the ticked sandboxes. Like every confirmation, it lists the safe choice first.
func deleteSandboxesBehaviour() screenBehaviour {
	return screenBehaviour{
		list: &listBehaviour{
			// No, go back; then Yes, delete.
			itemCount: func(State) int { return 2 },
			// Enter starts the deletion on the confirming choice, which blocks
			// every key but Ctrl+C until it finishes. "No, go back" returns to
			// Manage sandboxes with the ticks kept and Delete selected selected.
			onSelectItem: func(s *State) tea.Cmd {
				if s.nav.selection == confirmItem {
					return s.sandboxes.beginDeleting()
				}
				return s.goBackToParentScreen()
			},
		},
		// Manage sandboxes stays open underneath, so the listing and ticks
		// this confirmation lists are neither stopped nor restarted.
		parent: new(config.ManageSandboxesScreen),
		// The view lists the ticked sandboxes and hides the help line while
		// they are being deleted.
		view: func(s State, helpWidth int) string {
			listing := s.sandboxes
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
		},
		help: helpBehaviour{
			hasHelp: true,
		},
		isBusy: func(s State) bool { return s.sandboxes.isDeleting },
	}
}
