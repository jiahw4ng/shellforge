package app

import (
	"fmt"
	"shellforge/internal/config"
	"shellforge/internal/logging"
	"shellforge/internal/screens"
	"shellforge/internal/terminal"
	"shellforge/internal/ui"
	"slices"

	tea "charm.land/bubbletea/v2"
)

// terminalProblemHintFormat tells the learner how to leave a startup failure.
const terminalProblemHintFormat = "Then press %s to go back and try again."

// settingNameSuffix separates a setting's name from its current value.
const settingNameSuffix = ":"

// View renders the active screen inside Shellforge's application frame.
func (s State) View() tea.View {
	return s.applicationView(s.activeScreenView())
}

func (s State) activeScreenView() string {
	helpWidth := s.helpContentWidth()
	switch s.nav.screen {
	case config.LessonsScreen:
		return screens.LessonList(s.lessons.available, s.lessons.completed, s.nav.selection, s.lessons.err, s.navigationHelp(helpWidth))
	case config.LessonTerminalScreen:
		return s.lessonView()
	case config.SettingsScreen:
		return screens.Settings(s.settingsItems(), s.nav.selection, s.settings.message, s.settings.failed, s.navigationHelp(helpWidth))
	case config.CursorStyleScreen:
		return screens.CursorStyle(cursorStyleLabels(), s.currentCursorStyleIndex(), s.nav.selection, s.navigationHelp(helpWidth))
	case config.ThemeScreen:
		return screens.Theme(themeLabels(), s.currentThemeIndex(), s.nav.selection, s.navigationHelp(helpWidth))
	case config.ResetConfirmationScreen:
		return s.resetConfirmationView(helpWidth)
	case config.ResetSettingsScreen:
		return screens.ResetSettingsConfirmation(config.ResetSettingsItems, s.nav.selection, s.navigationHelp(helpWidth))
	case config.ManageSandboxesScreen:
		return s.manageSandboxesView(helpWidth)
	case config.SandboxTerminalScreen:
		return s.sandboxView()
	default:
		return screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
	}
}

// resetConfirmationView hides the help line while a reset is in progress.
func (s State) resetConfirmationView(helpWidth int) string {
	help := ""
	if !s.settings.isResetting {
		help = s.navigationHelp(helpWidth)
	}
	return screens.ResetConfirmation(config.ResetConfirmationItems, s.nav.selection, s.settings.isResetting, help)
}

// manageSandboxesView shows the latest sandbox listing, its spinner while it
// runs, or why it failed.
func (s State) manageSandboxesView(helpWidth int) string {
	listing := s.settings.sandboxes
	return screens.ManageSandboxes{
		Sandboxes: listing.found,
		ListedAt:  listing.listedAt,
		Loading:   listing.isLoading,
		Spinner:   listing.spinner.View(),
		Problem:   sandboxListingProblem(listing.err),
		Selection: s.nav.selection,
		Help:      s.navigationHelp(helpWidth),
	}.View()
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

func (s State) applicationView(content string) tea.View {
	if s.viewport.width <= 0 || s.viewport.height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(s.framed(content))
	view.AltScreen = true
	if s.term.session != nil {
		view.Cursor = s.positionedTerminalCursor(s.term.session.Cursor())
	}
	return view
}

// framed surrounds content with the application border. The lesson terminal
// screen also joins the border to the divider between instructions and shell.
func (s State) framed(content string) string {
	if s.nav.screen != config.LessonTerminalScreen {
		return ui.WithAppFrame(content, s.viewport.width, s.viewport.height)
	}
	contentWidth, _ := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	dividerColumn := screens.LessonInstructionWidth(contentWidth)
	return ui.WithAppFrameJunction(content, s.viewport.width, s.viewport.height, dividerColumn)
}

// settingsItems labels the preference items with the current choices, padding
// each "Name:" to the widest one so the values start in the same column.
func (s State) settingsItems() []string {
	items := slices.Clone(config.SettingsItems)
	values := map[int]string{
		settingsCursorStyleIndex: s.preferences.cursorStyle().Label,
		settingsThemeIndex:       s.themeSettingLabel(),
	}
	nameWidth := 0
	for index := range values {
		nameWidth = max(nameWidth, len(items[index])+len(settingNameSuffix))
	}
	for index, value := range values {
		items[index] = fmt.Sprintf("%-*s %s", nameWidth, items[index]+settingNameSuffix, value)
	}
	return items
}

// themeSettingLabel names the theme mode in Settings. Auto also shows the
// theme it currently resolves to.
func (s State) themeSettingLabel() string {
	if s.preferences.themeExplicit {
		return s.preferences.theme().Label
	}
	return ui.AutoThemeLabel + " (" + s.preferences.theme().Label + ")"
}

// themeLabels lists Auto first, then every theme, matching Theme screen selection.
func themeLabels() []string {
	labels := []string{ui.AutoThemeLabel}
	for _, theme := range ui.Themes {
		labels = append(labels, theme.Label)
	}
	return labels
}

func cursorStyleLabels() []string {
	labels := make([]string, len(ui.CursorStyles))
	for index, style := range ui.CursorStyles {
		labels[index] = style.Label
	}
	return labels
}

// positionedTerminalCursor moves the session's cursor into application
// coordinates and applies the learner's cursor style.
func (s State) positionedTerminalCursor(cursor *tea.Cursor) *tea.Cursor {
	if cursor == nil || !s.usesTerminal() || s.viewport.width <= 0 || s.viewport.height <= 0 {
		return nil
	}

	contentWidth, contentHeight := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	terminalWidth, terminalHeight := s.terminalDimensions()
	positioned := *cursor
	style := s.preferences.cursorStyle()
	positioned.Shape = style.Shape
	positioned.Blink = style.Blink
	positioned.X += (s.viewport.width-contentWidth)/2 + contentWidth - terminalWidth
	positioned.Y += (s.viewport.height-contentHeight)/2 + contentHeight - terminalHeight
	return &positioned
}

func (s State) sandboxView() string {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	return screens.SandboxTerminalPane(s.sandboxTerminalContent(), width, height)
}

// sandboxTerminalContent is what fills the standalone sandbox: the live
// terminal, the startup spinner, or the reason the terminal is not running.
func (s State) sandboxTerminalContent() string {
	if s.term.session != nil {
		return s.term.session.View()
	}
	if s.term.isStarting {
		return screens.TerminalStart(nil, s.term.startSteps, s.term.spinner.View())
	}
	if content, ok := s.terminalStoppedView(); ok {
		return content
	}
	return screens.TerminalFailure(s.term.output, s.term.err)
}

// lessonTerminalContent is what the lesson pane passes down for its terminal:
// the live terminal or the reason it is not running. It is empty while the
// terminal starts, so the screen can show its own loading state.
func (s State) lessonTerminalContent() string {
	if s.term.session != nil {
		return s.term.session.View()
	}
	content, _ := s.terminalStoppedView()
	return content
}

// terminalStoppedView explains why a terminal that is not running stopped:
// whatever it printed before exiting, otherwise a recognised startup problem.
// ok is false when there is nothing specific to show.
func (s State) terminalStoppedView() (string, bool) {
	if s.term.output != "" {
		return screens.TerminalFailure(s.term.output, s.term.err), true
	}
	return s.terminalProblemView()
}

// terminalProblemView explains a recognised startup failure with a summary,
// a remedy, and how to leave. ok is false when there is no start failure or it
// is one with no specific advice, so the caller shows the error text as is.
func (s State) terminalProblemView() (string, bool) {
	if s.term.err == nil || s.term.output != "" {
		return "", false
	}
	message, ok := terminal.Describe(s.term.err)
	if !ok {
		return "", false
	}
	return screens.TerminalProblem{
		Summary: message.Summary,
		Remedy:  message.Remedy,
		Hint:    fmt.Sprintf(terminalProblemHintFormat, config.Keys.SelectItem.Help().Key),
		Detail:  message.Detail,
		LogPath: logging.Path(),
	}.View(), true
}

func (s State) helpContentWidth() int {
	width, _ := ui.ApplicationContentDimensions(ui.DimensionWithFallback(s.viewport.width, 80), 1)
	return width
}

func (s State) lessonView() string {
	return s.lessonScreen().View()
}

func (s State) lessonScreen() screens.Lesson {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	instructionWidth := screens.LessonInstructionWidth(ui.DimensionWithFallback(width, 80))
	lesson := s.lessons.available[s.lessons.activeIdx]
	pageIndex := s.lessons.activePage
	return screens.Lesson{
		Lesson:             &lesson,
		PageIndex:          pageIndex,
		Guide:              s.lessons.guide,
		RevealedHints:      s.lessons.revealedHints,
		HintsExpanded:      s.lessons.hintsExpanded,
		TerminalContent:    s.lessonTerminalContent(),
		TerminalError:      s.term.err,
		TerminalLoading:    s.term.spinner.View(),
		TerminalSteps:      s.term.startSteps,
		TerminalStarting:   s.term.isStarting,
		AssertionResults:   s.lessons.progress.results,
		AssertionsChecking: s.lessons.progress.isChecking,
		AssertionsChecked:  s.lessons.progress.hasChecked,
		AssertionsExpanded: s.lessons.progress.assertionsExpanded,
		Help:               s.lessonHelp(instructionWidth),
		Width:              width,
		Height:             height,
	}
}

func (s *State) prepareLessonGuide() {
	if s.nav.screen != config.LessonTerminalScreen {
		return
	}
	s.lessons.guide = s.lessonScreen().PrepareGuide()
}
