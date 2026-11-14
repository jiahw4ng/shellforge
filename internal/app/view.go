package app

import (
	"shellforge/internal/config"
	"shellforge/internal/screens"
	"shellforge/internal/ui"

	tea "charm.land/bubbletea/v2"
)

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
	case config.DeleteSandboxesScreen:
		return s.deleteSandboxesView(helpWidth)
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

func (s State) helpContentWidth() int {
	width, _ := ui.ApplicationContentDimensions(ui.DimensionWithFallback(s.viewport.width, 80), 1)
	return width
}
