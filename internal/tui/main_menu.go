package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	case ptyStartedMsg:
		m.handlePTYStarted(msg)
		if m.pty != nil {
			return m, readPTYOutput(m.pty)
		}
	case ptyOutputMsg:
		m.appendPTYOutput(msg)
		return m, readPTYOutput(msg.session)
	case ptyClosedMsg:
		m.handlePTYClosed(msg)
	case tea.KeyMsg:
		if m.screen == ptyScreen {
			return m.handlePTYKey(msg)
		}

		shouldQuit := m.handleKeyMsg(msg)
		if shouldQuit {
			return m, tea.Quit
		}
		if m.screen == ptyScreen {
			return m, startPTY(m.width, m.height)
		}
	}

	return m, nil
}

// handleResize updates the model's width and height based on the window size message
func (m *Model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	if m.pty != nil {
		_ = m.pty.Resize(msg.Width, msg.Height)
	}
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
				m.screen = ptyScreen
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
func (m Model) View() string {
	var content string
	switch m.screen {
	case featureScreen:
		content = featureView()
	case ptyScreen:
		content = ptyView(m.ptyOutput, m.ptyError)
	default:
		content = menuView(m.selected)
	}

	if m.width <= 0 || m.height <= 0 {
		return content
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
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
