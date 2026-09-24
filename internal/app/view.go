package app

import (
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the active screen inside Shellforge's application frame.
func (m State) View() tea.View {
	var content string
	switch m.nav.screen {
	case lessonsScreen:
		content = screens.LessonList(m.lessons.available, m.lessons.completed, m.nav.selection, m.lessons.err)
	case lessonScreen:
		content = m.lessonView()
	case settingsScreen:
		content = screens.Settings(settingsItems, m.nav.selection, m.settings.message, m.settings.failed)
	case resetConfirmationScreen:
		content = screens.ResetConfirmation(resetConfirmationItems, m.nav.selection, m.settings.isResetting)
	case terminalScreen:
		if m.term.session != nil {
			content = m.term.session.View()
		} else if m.term.isStarting {
			content = screens.TerminalStart(nil, m.term.spinner.View())
		} else {
			content = screens.TerminalFailure(m.term.output, m.term.err)
		}
	default:
		content = screens.MainMenu(menuItems, m.nav.selection)
	}

	if m.viewport.width <= 0 || m.viewport.height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(ui.WithAppFrame(content, m.viewport.width, m.viewport.height))
	view.AltScreen = true
	return view
}

func (m State) lessonView() string {
	if m.lessons.activeIdx >= len(m.lessons.available) {
		return screens.LessonList(m.lessons.available, m.lessons.completed, m.nav.selection, m.lessons.err)
	}

	width, height := ui.ApplicationContentDimensions(m.viewport.width, m.viewport.height)
	terminalContent := ""
	if m.term.session != nil {
		terminalContent = m.term.session.View()
	} else if m.term.output != "" {
		terminalContent = screens.TerminalFailure(m.term.output, m.term.err)
	}
	lesson := m.lessons.available[m.lessons.activeIdx]
	pageIndex := m.lessons.activePage
	if pageIndex < 0 || pageIndex >= len(lesson.Pages) {
		pageIndex = 0
	}
	return screens.Lesson(screens.LessonRenderParams{
		Lesson:             &lesson,
		Page:               &lesson.Pages[pageIndex],
		PageIndex:          pageIndex,
		TerminalContent:    terminalContent,
		TerminalError:      m.term.err,
		TerminalLoading:    m.term.spinner.View(),
		TerminalStarting:   m.term.isStarting,
		AssertionResults:   m.lessons.progress.results,
		AssertionsChecking: m.lessons.progress.isChecking,
		AssertionsChecked:  m.lessons.progress.hasChecked,
		Width:              width,
		Height:             height,
	})
}
