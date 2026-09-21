package app

import (
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the active screen inside Shellforge's application frame.
func (m State) View() tea.View {
	var content string
	switch m.Nav.Screen {
	case lessonsScreen:
		content = screens.LessonList(m.Lessons.Available, m.Lessons.Completed, m.Nav.Selection, m.Lessons.Error)
	case lessonScreen:
		content = m.lessonView()
	case settingsScreen:
		content = screens.Settings(settingsItems, m.Nav.Selection, m.Settings.Message, m.Settings.Failed)
	case resetConfirmationScreen:
		content = screens.ResetConfirmation(resetConfirmationItems, m.Nav.Selection, m.Settings.isResetting)
	case terminalScreen:
		if m.Term.Session != nil {
			content = m.Term.Session.View()
		} else if m.Term.isStarting {
			content = screens.TerminalStart(nil, m.Term.Spinner.View())
		} else {
			content = screens.TerminalFailure(m.Term.Output, m.Term.Error)
		}
	default:
		content = screens.MainMenu(menuItems, m.Nav.Selection)
	}

	if m.Viewport.Width <= 0 || m.Viewport.Height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(ui.WithAppFrame(content, m.Viewport.Width, m.Viewport.Height))
	view.AltScreen = true
	return view
}

func (m State) lessonView() string {
	if m.Lessons.ActiveIndex >= len(m.Lessons.Available) {
		return screens.LessonList(m.Lessons.Available, m.Lessons.Completed, m.Nav.Selection, m.Lessons.Error)
	}

	width, height := ui.ApplicationContentDimensions(m.Viewport.Width, m.Viewport.Height)
	terminalContent := ""
	if m.Term.Session != nil {
		terminalContent = m.Term.Session.View()
	} else if m.Term.Output != "" {
		terminalContent = screens.TerminalFailure(m.Term.Output, m.Term.Error)
	}
	lesson := m.Lessons.Available[m.Lessons.ActiveIndex]
	pageIndex := m.Lessons.ActivePage
	if pageIndex < 0 || pageIndex >= len(lesson.Pages) {
		pageIndex = 0
	}
	return screens.Lesson(screens.LessonRenderParams{
		Lesson:             &lesson,
		Page:               &lesson.Pages[pageIndex],
		PageIndex:          pageIndex,
		TerminalContent:    terminalContent,
		TerminalError:      m.Term.Error,
		TerminalLoading:    m.Term.Spinner.View(),
		TerminalStarting:   m.Term.isStarting,
		AssertionResults:   m.Lessons.Progress.Results,
		AssertionsChecking: m.Lessons.Progress.isChecking,
		AssertionsChecked:  m.Lessons.Progress.hasChecked,
		Width:              width,
		Height:             height,
	})
}
