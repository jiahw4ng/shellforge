package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// New creates the initial main-menu model
func New() Model {
	return Model{}
}

// Init initializes the TUI model and returns a command to run
// in this case, it returns nil since no initialization is needed
func (Model) Init() tea.Cmd {
	return nil
}

// Update handles messages sent to the TUI model and updates the model's state accordingly
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg)
		if m.terminal != nil {
			return m.updateTerminal(msg)
		}
	case terminalStartedMsg:
		m.handleTerminalStarted(msg)
		if m.terminal != nil {
			return m, tea.Batch(m.terminal.Init(), waitForTerminalExit(m.terminalExit))
		}
	case terminalExitedMsg:
		m.closeTerminal()
		m.currentScreen = menuScreen
	case tea.KeyPressMsg:
		if m.currentScreen == terminalScreen {
			// user currently in the terminal screen
			if m.terminal != nil {
				// Bubbleterm handles its own key events,
				// so route the key message to the terminal
				return m.updateTerminal(msg)
			}
			if msg.String() == "ctrl+c" {
				// quit from the terminal screen
				return m, tea.Quit
			}
			if msg.String() == "enter" && m.terminalStartErr != nil {
				// if the terminal failed to start, return to the main menu
				m.currentScreen = menuScreen
				m.terminalStartErr = nil
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
		if m.currentScreen == terminalScreen {
			return m, startTerminal(m.termWidth, m.termHeight)
		}
	default:
		// Bubbleterm emits its own output messages. They are intentionally
		// unexported by the library, so route every remaining event to the
		// active terminal instead of trying to type-match those messages here.
		if m.currentScreen == terminalScreen && m.terminal != nil {
			return m.updateTerminal(message)
		}
	}

	return m, nil
}

// handleResize updates the model's width and height based on the window size message
func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.termWidth = msg.Width
	m.termHeight = msg.Height
}

// handleKeyMsg processes key messages and updates the model's state accordingly
// if returns true, the caller should quit the application
func (m *Model) handleKeyMsg(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.currentScreen == menuScreen && m.selectedOption > 0 {
			m.selectedOption--
		}
		if m.currentScreen == lessonsScreen && m.selectedLesson > 0 {
			m.selectedLesson--
		}
	case "down":
		if m.currentScreen == menuScreen && m.selectedOption < len(menuItems)-1 {
			m.selectedOption++
		}
		if m.currentScreen == lessonsScreen && m.selectedLesson < len(lessonItems) {
			m.selectedLesson++
		}
	case "enter":
		switch m.currentScreen {
		case menuScreen:
			switch m.selectedOption {
			case 0:
				m.currentScreen = terminalScreen
			case 1:
				m.currentScreen = lessonsScreen
			default:
				m.currentScreen = featureScreen
				m.featureReturnTo = menuScreen
			}
		case lessonsScreen:
			if m.selectedLesson == len(lessonItems) {
				m.currentScreen = menuScreen
			} else {
				m.currentScreen = featureScreen
				m.featureReturnTo = lessonsScreen
			}
		case featureScreen:
			m.currentScreen = m.featureReturnTo
		default:
			m.currentScreen = menuScreen
		}
	}

	return false
}

// View renders the TUI model's current state as a string
func (m Model) View() tea.View {
	var content string
	switch m.currentScreen {
	case lessonsScreen:
		content = displayLessonsView(m.selectedLesson)
	case featureScreen:
		content = displayFeatureView()
	case terminalScreen:
		if m.terminal != nil {
			view := m.terminal.View()
			view.AltScreen = true
			return view
		}
		content = terminalStartView(m.terminalStartErr)
	default:
		content = displayMenuViewWithSelectArrow(m.selectedOption)
	}

	if m.termWidth <= 0 || m.termHeight <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Left, lipgloss.Top, content))
	view.AltScreen = true
	return view
}

// displayMenuViewWithSelectArrow generates the view for the main menu screen, highlighting the currently selected item
func displayMenuViewWithSelectArrow(selection int) string {
	lines := []string{titleStyle.Render("Welcome to Shellforge!"), ""}
	for index, item := range menuItems {
		prefix := "  "
		if index == selection {
			prefix = "> "
			item = selected.Render(item)
		}
		lines = append(lines, prefix+item)
	}

	lines = append(lines, "", muted.Render("Use ↑/↓ to choose and Enter to continue. Ctrl+C exits."))
	return strings.Join(lines, "\n")
}
