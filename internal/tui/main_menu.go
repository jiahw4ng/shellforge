package tui

import (
	"shellforge/internal/ui"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// New creates the initial main-menu model
func New() State {
	return State{}
}

// Init initializes the TUI model and returns a command to run
// in this case, it returns nil since no initialization is needed
func (State) Init() tea.Cmd {
	return nil
}

// Update handles messages sent to the TUI model and updates the model's state accordingly
func (m State) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		if m.Terminal != nil {
			if m.CurrentScreen == lessonScreen {
				return m, m.resizeLessonTerminal()
			}
			return m, m.resizeSandboxTerminal()
		}
	case terminalStartedMsg:
		m.handleTerminalStarted(msg)
		if m.Terminal != nil {
			commands := []tea.Cmd{m.Terminal.Init(), waitForTerminalExit(m.TermExit)}
			if m.CurrentScreen == lessonScreen {
				commands = append(commands, m.resizeLessonTerminal())
			}
			return m, tea.Batch(commands...)
		}
	case terminalExitedMsg:
		m.closeTerminal()
		if m.CurrentScreen == lessonScreen {
			m.CurrentScreen = lessonsScreen
		} else {
			m.CurrentScreen = menuScreen
		}
	case tea.KeyPressMsg:
		if m.usesTerminal() {
			// user currently in the terminal screen
			if m.Terminal != nil {
				// Bubbleterm handles its own key events,
				// so route the key message to the terminal
				return m.updateTerminal(msg)
			}
			if msg.String() == "ctrl+c" {
				// quit from the terminal screen
				return m, tea.Quit
			}
			if msg.String() == "enter" && m.TermState.TermErr != nil {
				// If the terminal failed to start, return to its parent menu.
				if m.CurrentScreen == lessonScreen {
					m.CurrentScreen = lessonsScreen
				} else {
					m.CurrentScreen = menuScreen
				}
				m.TermState.TermErr = nil
			}
			return m, nil
		}

		// User is in the main menu, lesson menu, or a placeholder feature screen.
		shouldQuit := m.handleKeyMsg(msg)
		if shouldQuit {
			return m, tea.Quit
		}

		// if handling the key press resulted in a screen change to the terminal,
		// start the terminal
		if m.usesTerminal() {
			width, height := m.terminalDimensions()
			return m, startTerminal(width, height)
		}
	default:
		// Bubbleterm emits its own output messages. They are intentionally
		// unexported by the library, so route every remaining event to the
		// active terminal instead of trying to type-match those messages here.
		if m.usesTerminal() && m.Terminal != nil {
			return m.updateTerminal(message)
		}
	}

	return m, nil
}

// handleResize updates the model's width and height based on the window size message
func (m *State) handleResize(msg tea.WindowSizeMsg) {
	m.TermWidth = msg.Width
	m.TermHeight = msg.Height
}

// usesTerminal reports whether the active screen embeds a Bubbleterm session.
func (m State) usesTerminal() bool {
	return m.CurrentScreen == terminalScreen || m.CurrentScreen == lessonScreen
}

// resizeLessonTerminal resizes Bubbleterm to the right-hand lesson pane rather
// than forwarding the full outer-terminal dimensions to it.
func (m State) resizeLessonTerminal() tea.Cmd {
	width, height := lessonTerminalDimensions(ui.ApplicationContentDimensions(m.TermWidth, m.TermHeight))
	return m.Terminal.Resize(width, height)
}

// resizeSandboxTerminal resizes a full-screen sandbox to the area inside the
// application border.
func (m State) resizeSandboxTerminal() tea.Cmd {
	width, height := ui.ApplicationContentDimensions(m.TermWidth, m.TermHeight)
	return m.Terminal.Resize(width, height)
}

// terminalDimensions returns the correct Bubbleterm size for the active
// screen before its sandbox process is started.
func (m State) terminalDimensions() (int, int) {
	width, height := ui.ApplicationContentDimensions(m.TermWidth, m.TermHeight)
	if m.CurrentScreen == lessonScreen {
		return lessonTerminalDimensions(width, height)
	}
	return terminalDimension(width, 80), terminalDimension(height, 24)
}

// handleKeyMsg processes key messages and updates the model's state accordingly
// if returns true, the caller should quit the application
func (m *State) handleKeyMsg(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.CurrentScreen == menuScreen && m.SelectedOption > 0 {
			m.SelectedOption--
		}
		if m.CurrentScreen == lessonsScreen && m.SelectedLesson > 0 {
			m.SelectedLesson--
		}
	case "down":
		if m.CurrentScreen == menuScreen && m.SelectedOption < len(menuItems)-1 {
			m.SelectedOption++
		}
		if m.CurrentScreen == lessonsScreen && m.SelectedLesson < len(lessonItems) {
			m.SelectedLesson++
		}
	case "enter":
		switch m.CurrentScreen {
		case menuScreen:
			switch m.SelectedOption {
			case 0:
				m.CurrentScreen = terminalScreen
			case 1:
				m.CurrentScreen = lessonsScreen
			default:
				m.CurrentScreen = featureScreen
				m.FeatureReturnTo = menuScreen
			}
		case lessonsScreen:
			if m.SelectedLesson == len(lessonItems) {
				m.CurrentScreen = menuScreen
			} else {
				m.ActiveLesson = m.SelectedLesson
				m.CurrentScreen = lessonScreen
			}
		case featureScreen:
			m.CurrentScreen = m.FeatureReturnTo
		default:
			m.CurrentScreen = menuScreen
		}
	}

	return false
}

// View renders the TUI model's current state as a string
func (m State) View() tea.View {
	var content string
	switch m.CurrentScreen {
	case lessonsScreen:
		content = displayLessonsView(m.SelectedLesson)
	case lessonScreen:
		innerWidth, innerHeight := ui.ApplicationContentDimensions(m.TermWidth, m.TermHeight)
		terminalContent := ""
		if m.Terminal != nil {
			terminalContent = m.Terminal.View().Content
		}
		content = displayLessonView(m.ActiveLesson, terminalContent, m.TermErr, innerWidth, innerHeight)
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

	view := tea.NewView(ui.WithDisplayApplicationFrame(content, m.TermWidth, m.TermHeight))
	view.AltScreen = true
	return view
}

// displayMenuViewWithSelectArrow generates the view for the main menu screen, highlighting the currently selected item
func displayMenuViewWithSelectArrow(selection int) string {
	lines := []string{ui.TitleStyle.Render("Welcome to Shellforge!"), ""}
	for index, item := range menuItems {
		prefix := "  "
		if index == selection {
			prefix = "> "
			item = ui.SelectedStyle.Render(item)
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "", ui.MutedStyle.Render("Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."))
	return strings.Join(lines, "\n")
}
