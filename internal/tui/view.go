package tui

import (
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

// View renders the TUI model's current state as a string
func (m State) View() tea.View {
	var content string
	switch m.CurrentScreen {
	case lessonsScreen:
		content = displayLessonsView(m.Lessons, m.SelectedLesson, m.LessonErr)
	case lessonScreen:
		innerWidth, innerHeight := ui.ApplicationContentDimensions(m.TermWidth, m.TermHeight)
		terminalContent := ""
		if m.Terminal != nil {
			terminalContent = m.Terminal.View().Content
		}
		if m.ActiveLesson < len(m.Lessons) {
			content = displayLessonView(m.Lessons[m.ActiveLesson], terminalContent, m.TermErr, innerWidth, innerHeight)
		} else {
			content = displayLessonsView(m.Lessons, m.SelectedLesson, m.LessonErr)
		}
	case featureScreen:
		content = displayFeatureView()
	case terminalScreen:
		if m.Terminal != nil {
			content = m.Terminal.View().Content
		} else {
			content = terminalStartView(m.TermErr)
		}
	default:
		content = displayMenuViewWithSelectArrow(m.SelectedOption)
	}

	if m.TermWidth <= 0 || m.TermHeight <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(ui.WithAppFrame(content, m.TermWidth, m.TermHeight))
	view.AltScreen = true
	return view
}
