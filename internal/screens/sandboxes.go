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
// It is read-only: Back is its only selectable item.
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
	// Selection is the selected item; Back is item 0.
	Selection int
	// Help is the rendered navigation help.
	Help string
}

// View renders the title, the loading line, problem, or sandbox rows, and Back.
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
		lines = append(lines, m.rows()...)
	}
	lines = append(lines,
		"",
		selectableItem(0, m.Selection, config.BackMenuLabel),
		"",
		m.Help,
	)
	return strings.Join(lines, "\n")
}

// rows aligns each sandbox's name, state, and age into columns.
func (m ManageSandboxes) rows() []string {
	nameWidth, stateWidth := 0, 0
	for _, sandbox := range m.Sandboxes {
		nameWidth = max(nameWidth, len(sandbox.Name))
		stateWidth = max(stateWidth, len(sandbox.State))
	}
	rows := make([]string, len(m.Sandboxes))
	for index, sandbox := range m.Sandboxes {
		row := fmt.Sprintf("%-*s  %-*s  %s", nameWidth, sandbox.Name, stateWidth, sandbox.State, age(sandbox.CreatedAt, m.ListedAt))
		rows[index] = unselectedItemPrefix + ui.TextStyle.Render(row)
	}
	return rows
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
