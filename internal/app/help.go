package app

import (
	"shellforge/internal/config"
	"shellforge/internal/ui"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
)

// Warnings shown above the help line when progress cannot be persisted.
const (
	storageUnavailableNotice = "Progress and settings can't be saved. They'll be lost when you quit."
	progressNotSavedNotice   = "Your lesson progress couldn't be saved. It will be lost when you quit."
	progressNotLoadedNotice  = "Saved progress couldn't be loaded, so some completed lessons may not show a checkmark."
)

func newHelpModel() help.Model {
	model := help.New()
	applyHelpStyles(&model)
	return model
}

// applyHelpStyles copies the active theme's help style into model. Call it again
// after a theme change, because the model keeps its own copies.
func applyHelpStyles(model *help.Model) {
	model.Styles.Ellipsis = ui.HelpStyle
	model.Styles.ShortKey = ui.HelpStyle
	model.Styles.ShortDesc = ui.HelpStyle
	model.Styles.ShortSeparator = ui.HelpStyle
	model.Styles.FullKey = ui.HelpStyle
	model.Styles.FullDesc = ui.HelpStyle
	model.Styles.FullSeparator = ui.HelpStyle
}

// navigationHelp renders toggleable shortcuts shared by non-terminal menu screens.
func (s State) navigationHelp(width int) string {
	return s.renderHelp(width, config.NavigationHelpKeyBindingGroups)
}

// storageNotice returns the warning to show above the help line, or an empty
// string while persistence is working. Only the most serious applicable
// problem is reported: unavailable, then not saved, then not loaded.
func (s State) storageNotice() string {
	switch {
	case s.lessons.store == nil:
		return storageUnavailableNotice
	case s.lessons.progressNotSaved:
		return progressNotSavedNotice
	case s.lessons.progressNotLoaded:
		return progressNotLoadedNotice
	default:
		return ""
	}
}

// renderHelp renders the help disclosure for a screen. Every screen with a help
// line goes through here, so the storage notice appears directly above it.
func (s State) renderHelp(width int, groups [][]key.Binding) string {
	helpModel := s.help
	helpModel.SetWidth(max(width-2, 1))
	body := helpModel.FullHelpView(groups)
	disclosure := ui.Disclosure(ui.DisclosureOptions{
		Binding:   config.Keys.ToggleHelp,
		Expanded:  s.help.ShowAll,
		Label:     config.HelpDisclosureLabel,
		Body:      body,
		EmptyBody: ui.HelpStyle.Render("No help available."),
		Width:     width,
	})
	notice := s.storageNotice()
	if notice == "" {
		return disclosure
	}
	return ui.WarningStyle.Render(ansi.Wrap(notice, max(width, 1), "")) + "\n" + disclosure
}
