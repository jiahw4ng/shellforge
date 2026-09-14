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
		m.screen = menuScreen
	case tea.KeyPressMsg:
		if m.screen == terminalScreen {
			if m.terminal != nil {
				return m.updateTerminal(msg)
			}
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "enter" && m.terminalStartErr != nil {
				m.screen = menuScreen
				m.terminalStartErr = nil
			}
			return m, nil
		}

		shouldQuit := m.handleKeyMsg(msg)
		if shouldQuit {
			return m, tea.Quit
		}
		if m.screen == terminalScreen {
			return m, startTerminal(m.width, m.height)
		}
	default:
		// Bubbleterm emits its own output messages. They are intentionally
		// unexported by the library, so route every remaining event to the
		// active terminal instead of trying to type-match those messages here.
		if m.screen == terminalScreen && m.terminal != nil {
			return m.updateTerminal(message)
		}
	}

	return m, nil
}

// handleResize updates the model's width and height based on the window size message
func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
}

// handleKeyMsg processes key messages and updates the model's state accordingly
func (m *Model) handleKeyMsg(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "ctrl+c":
		return true
	case "up":
		if m.screen == menuScreen && m.selected > 0 {
			m.selected--
		}
	case "down":
		if m.screen == menuScreen && m.selected < len(menuItems)-1 {
			m.selected++
		}
	case "enter":
		if m.screen == menuScreen {
			if m.selected == 0 {
				m.screen = terminalScreen
			} else {
				m.screen = featureScreen
			}
		} else {
			m.screen = menuScreen
		}
	}

	return false
}

// View renders the TUI model's current state as a string
func (m Model) View() tea.View {
	var content string
	switch m.screen {
	case featureScreen:
		content = featureView()
	case terminalScreen:
		if m.terminal != nil {
			view := m.terminal.View()
			view.AltScreen = true
			return view
		}
		content = terminalStartView(m.terminalStartErr)
	default:
		content = menuView(m.selected)
	}

	if m.width <= 0 || m.height <= 0 {
		view := tea.NewView(content)
		view.AltScreen = true
		return view
	}

	view := tea.NewView(lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content))
	view.AltScreen = true
	return view
}

// menuView generates the view for the main menu screen, highlighting the currently selected item
func menuView(selection int) string {
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
