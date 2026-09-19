package app

import (
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the active screen inside Shellforge's application frame.
func (m State) View() tea.View {
	var content string
	switch m.CurrentScreen {
	case lessonsScreen:
		content = screens.LessonList(m.Lessons, m.SelectedLesson, m.LessonErr)
	case lessonScreen:
		content = m.lessonView()
	case featureScreen:
		content = screens.Feature()
	case terminalScreen:
		if m.Terminal != nil {
			content = m.Terminal.View()
		} else if m.TerminalStarting {
			content = screens.TerminalStart(nil, m.TerminalSpinner.View())
		} else {
			content = screens.TerminalFailure(m.TerminalOutput, m.TerminalErr)
		}
	default:
		content = screens.MainMenu(menuItems, m.SelectedOption)
	}

	if m.Width <= 0 || m.Height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(ui.WithAppFrame(content, m.Width, m.Height))
	view.AltScreen = true
	return view
}

func (m State) lessonView() string {
	if m.ActiveLesson >= len(m.Lessons) {
		return screens.LessonList(m.Lessons, m.SelectedLesson, m.LessonErr)
	}

	width, height := ui.ApplicationContentDimensions(m.Width, m.Height)
	terminalContent := ""
	if m.Terminal != nil {
		terminalContent = m.Terminal.View()
	} else if m.TerminalOutput != "" {
		terminalContent = screens.TerminalFailure(m.TerminalOutput, m.TerminalErr)
	}
	lesson := m.Lessons[m.ActiveLesson]
	pageIndex := m.ActivePage
	if pageIndex < 0 || pageIndex >= len(lesson.Pages) {
		pageIndex = 0
	}
	return screens.Lesson(screens.LessonRenderParams{
		Lesson:           &lesson,
		Page:             &lesson.Pages[pageIndex],
		PageIndex:        pageIndex,
		TerminalContent:  terminalContent,
		TerminalError:    m.TerminalErr,
		TerminalLoading:  m.TerminalSpinner.View(),
		TerminalStarting: m.TerminalStarting,
		Width:            width,
		Height:           height,
	})
}
