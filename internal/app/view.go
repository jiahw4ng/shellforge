package app

import (
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the active screen inside Shellforge's application frame.
func (s State) View() tea.View {
	var content string
	switch s.nav.screen {
	case lessonsScreen:
		content = screens.LessonList(s.lessons.available, s.lessons.completed, s.nav.selection, s.lessons.err, s.navigationHelp())
	case lessonScreen:
		content = s.lessonView()
	case settingsScreen:
		content = screens.Settings(settingsItems, s.nav.selection, s.settings.message, s.settings.failed, s.navigationHelp())
	case resetConfirmationScreen:
		helpView := s.navigationHelp()
		if s.settings.isResetting {
			helpView = ""
		}
		content = screens.ResetConfirmation(resetConfirmationItems, s.nav.selection, s.settings.isResetting, helpView)
	case terminalScreen:
		if s.term.session != nil {
			content = s.term.session.View()
		} else if s.term.isStarting {
			content = screens.TerminalStart(nil, s.term.spinner.View())
		} else {
			content = screens.TerminalFailure(s.term.output, s.term.err)
		}
	default:
		content = screens.MainMenu(menuItems, s.nav.selection, s.navigationHelp())
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
		TerminalContent:    terminalContent,
		TerminalError:      s.term.err,
		TerminalLoading:    s.term.spinner.View(),
		TerminalStarting:   s.term.isStarting,
		AssertionResults:   s.lessons.progress.results,
		AssertionsChecking: s.lessons.progress.isChecking,
		AssertionsChecked:  s.lessons.progress.hasChecked,
		Help:               s.lessonHelp(),
		Width:              width,
		Height:             height,
	}
}

func (s *State) prepareLessonGuide() {
	if s.nav.screen != lessonScreen || s.lessons.activeIdx < 0 || s.lessons.activeIdx >= len(s.lessons.available) {
		return
	}
	s.lessons.guide = screens.PrepareLessonGuide(s.getLessonRenderParams())
}
