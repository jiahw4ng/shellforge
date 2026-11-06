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
		helpView := s.navigationHelp(helpWidth)
		if s.settings.isResetting {
			helpView = ""
		}
		return screens.ResetConfirmation(config.ResetConfirmationItems, s.nav.selection, s.settings.isResetting, helpView)
	case config.SandboxTerminalScreen:
		return s.sandboxView()
	default:
		return screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
	}
}

func (s State) applicationView(content string) tea.View {
	if s.viewport.width <= 0 || s.viewport.height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	var framed string
	if s.nav.screen == config.LessonTerminalScreen {
		contentWidth, _ := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
		dividerColumn := screens.LessonInstructionWidth(contentWidth)
		framed = ui.WithAppFrameJunction(content, s.viewport.width, s.viewport.height, dividerColumn)
	} else {
		framed = ui.WithAppFrame(content, s.viewport.width, s.viewport.height)
	}

	view := tea.NewView(framed)
	view.AltScreen = true
	if s.term.session != nil {
		view.Cursor = s.positionedTerminalCursor(s.term.session.Cursor())
	}
	return view
}

// settingsItems labels the preference items with the current choices, padding
// each "Name:" to the widest one so the values start in the same column.
func (s State) settingsItems() []string {
	items := slices.Clone(config.SettingsItems)
	values := map[int]string{
		settingsCursorStyleIndex: s.preferences.cursorStyle().Label,
		settingsThemeIndex:       s.preferences.theme().Label,
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

func themeLabels() []string {
	labels := make([]string, len(ui.Themes))
	for index, theme := range ui.Themes {
		labels[index] = theme.Label
	}
	return labels
}

func cursorStyleLabels() []string {
	labels := make([]string, len(config.CursorStyles))
	for index, style := range config.CursorStyles {
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
	terminalContent := ""
	if s.term.session != nil {
		terminalContent = s.term.session.View()
	} else if s.term.isStarting {
		terminalContent = screens.TerminalStart(nil, s.term.spinner.View())
	} else if problem, ok := s.terminalProblemView(); ok {
		terminalContent = problem
	} else {
		terminalContent = screens.TerminalFailure(s.term.output, s.term.err)
	}
	return screens.TerminalPane(terminalContent, width, height)
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
		Hint:    fmt.Sprintf(terminalProblemHintFormat, config.Keys.ReturnBack.Help().Key),
		Detail:  message.Detail,
		LogPath: logging.Path(),
	}.View(), true
}

func (s State) helpContentWidth() int {
	width, _ := ui.ApplicationContentDimensions(ui.DimensionWithFallback(s.viewport.width, 80), 1)
	return width
}

func (s State) lessonView() string {
	return screens.Lesson(s.getLessonRenderParams())
}

func (s State) getLessonRenderParams() screens.LessonRenderParams {
	width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
	terminalContent := ""
	if s.term.session != nil {
		terminalContent = s.term.session.View()
	} else if s.term.output != "" {
		terminalContent = screens.TerminalFailure(s.term.output, s.term.err)
	} else if problem, ok := s.terminalProblemView(); ok {
		terminalContent = problem
	}
	lesson := s.lessons.available[s.lessons.activeIdx]
	pageIndex := s.lessons.activePage
	return screens.LessonRenderParams{
		Lesson:             &lesson,
		PageIndex:          pageIndex,
		Guide:              s.lessons.guide,
		RevealedHints:      s.lessons.revealedHints,
		HintsExpanded:      s.lessons.hintsExpanded,
		TerminalContent:    terminalContent,
		TerminalError:      s.term.err,
		TerminalLoading:    s.term.spinner.View(),
		TerminalStarting:   s.term.isStarting,
		AssertionResults:   s.lessons.progress.results,
		AssertionsChecking: s.lessons.progress.isChecking,
		AssertionsChecked:  s.lessons.progress.hasChecked,
		AssertionsExpanded: s.lessons.progress.assertionsExpanded,
		Help:               s.lessonHelp(screens.LessonInstructionWidth(ui.DimensionWithFallback(width, 80))),
		Width:              width,
		Height:             height,
	}
}

func (s *State) prepareLessonGuide() {
	if s.nav.screen != config.LessonTerminalScreen {
		return
	}
	s.lessons.guide = screens.PrepareLessonGuide(s.getLessonRenderParams())
}
