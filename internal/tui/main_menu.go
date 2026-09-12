package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// New creates the initial main-menu model
func New() SFModel {
	return SFModel{}
}

// Init initializes the TUI model and returns a command to run
// in this case, it returns nil since no initialization is needed
func (SFModel) Init() tea.Cmd {
	return nil
}

// Update handles messages sent to the TUI model and updates the model's state accordingly
func (m SFModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg)
	case tea.KeyMsg:
		shouldQuit := m.handleKeyMsg(msg)
		if shouldQuit {
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleResize updates the model's width and height based on the window size message
func (m *SFModel) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
}

// handleKeyMsg processes key messages and updates the model's state accordingly
func (m *SFModel) handleKeyMsg(msg tea.KeyMsg) bool {
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
			m.screen = featureScreen
		} else {
			m.screen = menuScreen
		}
	}

	return false
}

// View renders the TUI model's current state as a string
func (m SFModel) View() string {
	var content string
	if m.screen == featureScreen {
		content = featureView()
	} else {
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
