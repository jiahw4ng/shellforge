package screens

import (
	"fmt"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"shellforge/internal/ui"
	"strings"
	"time"
)

// ManageSandboxes is the Settings screen that lists every sandbox Shellforge
// created, including ones left behind by windows that closed without cleanup.
// When sandboxes are listed, its items are Select all, one row per sandbox,
// Delete selected, and Back; otherwise Back is its only item.
type ManageSandboxes struct {
	// Sandboxes is the latest listing, newest first.
	Sandboxes []container.ManagedSandbox
	// ListedAt is when the listing was taken; ages are measured from it.
	ListedAt time.Time
	// Loading reports that a listing is in progress.
	Loading bool
	// Spinner is the spinner frame shown while Loading.
	Spinner string
	// Problem replaces the listing when it failed. It is already rendered,
	// such as a TerminalProblem view for a Docker that is not running.
	Problem string
	// Chosen holds the IDs of the ticked sandboxes.
	Chosen map[string]bool
	// Selection is the selected item, counted as Item describes.
	Selection int
	// Message reports the latest deletion; Failed marks it as an error.
	Message string
	Failed  bool
	// Help is the rendered navigation help.
	Help string
}

// SandboxItemKind is what one selectable item on Manage sandboxes does.
type SandboxItemKind int

const (
	// SandboxBackItem returns to Settings.
	SandboxBackItem SandboxItemKind = iota
	// SandboxSelectAllItem ticks every sandbox, or unticks them all when all are ticked.
	SandboxSelectAllItem
	// SandboxRowItem ticks or unticks one sandbox.
	SandboxRowItem
	// SandboxDeleteItem deletes the ticked sandboxes after confirmation.
	SandboxDeleteItem
)

// Listed reports whether the screen shows sandbox rows, and so the Select all
// and Delete selected items around them.
func (m ManageSandboxes) Listed() bool {
	return !m.Loading && m.Problem == "" && len(m.Sandboxes) > 0
}

// ItemCount is the number of selectable items.
func (m ManageSandboxes) ItemCount() int {
	if !m.Listed() {
		return 1
	}
	return len(m.Sandboxes) + 3
}

// Item reports what the item at index does and, for a sandbox row, which
// sandbox it is. Items run Select all, the rows, Delete selected, then Back.
func (m ManageSandboxes) Item(index int) (SandboxItemKind, int) {
	if !m.Listed() {
		return SandboxBackItem, 0
	}
	switch {
	case index == 0:
		return SandboxSelectAllItem, 0
	case index <= len(m.Sandboxes):
		return SandboxRowItem, index - 1
	case index == len(m.Sandboxes)+1:
		return SandboxDeleteItem, 0
	default:
		return SandboxBackItem, 0
	}
}

// ChosenCount is how many listed sandboxes are ticked.
func (m ManageSandboxes) ChosenCount() int {
	count := 0
	for _, sandbox := range m.Sandboxes {
		if m.Chosen[sandbox.ID] {
			count++
		}
	}
	return count
}

// View renders the title, the loading line, problem, or sandbox items, and Back.
func (m ManageSandboxes) View() string {
	lines := []string{ui.TitleStyle.Render(config.SettingsItemManageSandboxes), ""}
	switch {
	case m.Loading:
		lines = append(lines, ui.MutedStyle.Render(listingSandboxesLine+" "+m.Spinner))
	case m.Problem != "":
		lines = append(lines, m.Problem)
	case len(m.Sandboxes) == 0:
		lines = append(lines, ui.MutedStyle.Render(noSandboxesLine))
	default:
		lines = append(lines, m.listedItems()...)
	}
	lines = append(lines, "", selectableItem(m.ItemCount()-1, m.Selection, config.BackMenuLabel))
	if m.Message != "" {
		style := ui.SuccessStyle
		if m.Failed {
			style = ui.FailureStyle
		}
		lines = append(lines, "", style.Render(m.Message))
	}
	lines = append(lines, "", m.Help)
	return strings.Join(lines, "\n")
}

// DeleteSandboxesConfirmation asks the learner to confirm deleting the ticked
// sandboxes. Running ones are marked, because another Shellforge window may be
// using them; deleting them is allowed.
type DeleteSandboxesConfirmation struct {
	// Sandboxes are the ticked sandboxes, in listing order.
	Sandboxes []container.ManagedSandbox
	// Selection is 0 for "No, go back" and 1 for the confirming choice.
	Selection int
	// Deleting reports that the deletion is in progress.
	Deleting bool
	// Spinner is the spinner frame shown while Deleting.
	Spinner string
	// Help is the rendered navigation help, empty while Deleting.
	Help string
}

// View renders the question, the sandboxes to delete, and the two choices.
func (m DeleteSandboxesConfirmation) View() string {
	lines := []string{ui.TitleStyle.Render(fmt.Sprintf(deleteSandboxesTitleFormat, sandboxCount(len(m.Sandboxes)))), ""}
	nameWidth := 0
	for _, sandbox := range m.Sandboxes {
		nameWidth = max(nameWidth, len(sandbox.Name))
	}
	for _, sandbox := range m.Sandboxes {
		line := unselectedItemPrefix + sandbox.Name
		if sandbox.State == runningState {
			line = fmt.Sprintf("%s%-*s  %s", unselectedItemPrefix, nameWidth, sandbox.Name, ui.WarningStyle.Render(runningSandboxWarning))
		}
		lines = append(lines, line)
	}
	lines = append(lines,
		"",
		cannotBeUndoneLine,
		"",
		selectableItem(0, m.Selection, deleteCancelLabel),
		selectableItem(1, m.Selection, fmt.Sprintf(deleteConfirmFormat, sandboxCount(len(m.Sandboxes)))),
	)
	if m.Deleting {
		lines = append(lines, "", ui.MutedStyle.Render(deletingSandboxesLine+" "+m.Spinner))
	}
	if m.Help != "" {
		lines = append(lines, "", m.Help)
	}
	return strings.Join(lines, "\n")
}

// sandboxCount names a number of sandboxes: "1 sandbox" or "3 sandboxes".
func sandboxCount(count int) string {
	if count == 1 {
		return "1 sandbox"
	}
	return fmt.Sprintf("%d sandboxes", count)
}

// listedItems renders Select all, the sandbox rows, and Delete selected.
// Delete selected is dimmed while nothing is ticked.
func (m ManageSandboxes) listedItems() []string {
	chosen := m.ChosenCount()
	lines := []string{selectableItem(0, m.Selection, checkbox(chosen == len(m.Sandboxes))+selectAllLabel)}
	for index, row := range m.rows() {
		lines = append(lines, selectableItem(index+1, m.Selection, row))
	}
	deleteIndex := len(m.Sandboxes) + 1
	deleteLabel := fmt.Sprintf(deleteSelectedFormat, chosen)
	if chosen == 0 && m.Selection != deleteIndex {
		deleteLabel = ui.MutedStyle.Render(deleteLabel)
	}
	return append(lines, "", selectableItem(deleteIndex, m.Selection, deleteLabel))
}

// rows aligns each sandbox's checkbox, name, state, and age into columns.
func (m ManageSandboxes) rows() []string {
	nameWidth, stateWidth := 0, 0
	for _, sandbox := range m.Sandboxes {
		nameWidth = max(nameWidth, len(sandbox.Name))
		stateWidth = max(stateWidth, len(sandbox.State))
	}
	rows := make([]string, len(m.Sandboxes))
	for index, sandbox := range m.Sandboxes {
		rows[index] = fmt.Sprintf("%s%-*s  %-*s  %s", checkbox(m.Chosen[sandbox.ID]), nameWidth, sandbox.Name, stateWidth, sandbox.State, age(sandbox.CreatedAt, m.ListedAt))
	}
	return rows
}

// checkbox renders a ticked or unticked box followed by a space.
func checkbox(ticked bool) string {
	if ticked {
		return tickedBox
	}
	return untickedBox
}

// age describes how long before now a sandbox was created, in the largest
// whole unit: "just now", "12m ago", "3h ago", or "2d ago".
func age(created, now time.Time) string {
	elapsed := now.Sub(created)
	switch {
	case elapsed < time.Minute:
		return justNowAge
	case elapsed < time.Hour:
		return fmt.Sprintf(ageFormat, int(elapsed/time.Minute), "m")
	case elapsed < 24*time.Hour:
		return fmt.Sprintf(ageFormat, int(elapsed/time.Hour), "h")
	default:
		return fmt.Sprintf(ageFormat, int(elapsed/(24*time.Hour)), "d")
	}
}
