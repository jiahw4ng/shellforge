package app

import (
	"shellforge/internal/config"
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the active screen inside Shellforge's application frame.
func (s State) View() tea.View {
	var content string
	helpWidth := s.helpContentWidth()
	switch s.nav.screen {
	case config.LessonsScreen:
		content = screens.LessonList(s.lessons.available, s.lessons.completed, s.nav.selection, s.lessons.err, s.navigationHelp(helpWidth))
	case config.LessonTerminalScreen:
		content = s.lessonView()
	case config.SettingsScreen:
		content = screens.Settings(config.SettingsItems, s.nav.selection, s.settings.message, s.settings.failed, s.navigationHelp(helpWidth))
	case config.ResetConfirmationScreen:
		helpView := s.navigationHelp(helpWidth)
		if s.settings.isResetting {
			helpView = ""
		}
		content = screens.ResetConfirmation(config.ResetConfirmationItems, s.nav.selection, s.settings.isResetting, helpView)
	case config.SandboxTerminalScreen:
		width, height := ui.ApplicationContentDimensions(s.viewport.width, s.viewport.height)
		terminalContent := ""
		if s.term.session != nil {
			terminalContent = s.term.session.View()
		} else if s.term.isStarting {
			terminalContent = screens.TerminalStart(nil, s.term.spinner.View())
		} else {
			terminalContent = screens.TerminalFailure(s.term.output, s.term.err)
		}
		content = screens.TerminalPane(terminalContent, width, height)
	default:
		content = screens.MainMenu(config.MenuItems, s.nav.selection, s.navigationHelp(helpWidth))
	}

	if s.viewport.width <= 0 || s.viewport.height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(ui.WithAppFrame(content, s.viewport.width, s.viewport.height))
	view.AltScreen = true
	return view
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
