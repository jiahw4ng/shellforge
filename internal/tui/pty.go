package tui

import (
	"regexp"
	"strings"

	shellpty "shellforge/internal/pty"

	tea "github.com/charmbracelet/bubbletea"
)

const maxPTYOutput = 32 * 1024

const clearScreenSequence = "\x1b[2J"

var (
	ansiCSI = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
	ansiOSC = regexp.MustCompile(`\x1b\][^\a]*(?:\a|\x1b\\)`)
	crlf    = regexp.MustCompile(`\r+\n`)
)

type ptyStartedMsg struct {
	session *shellpty.Session
	err     error
}

type ptyOutputMsg struct {
	session *shellpty.Session
	output  string
}

type ptyClosedMsg struct {
	session *shellpty.Session
}

func startPTY(width, height int) tea.Cmd {
	return func() tea.Msg {
		session, err := shellpty.Start(width, height)
		return ptyStartedMsg{session: session, err: err}
	}
}

func (m *Model) handlePTYStarted(message ptyStartedMsg) {
	if message.err != nil {
		m.ptyError = message.err
		return
	}

	m.pty = message.session
	m.ptyOutput = ""
	m.ptyError = nil
}

func (m *Model) appendPTYOutput(message ptyOutputMsg) {
	if m.pty != message.session {
		return
	}

	output := message.output
	if clearIndex := strings.LastIndex(output, clearScreenSequence); clearIndex >= 0 {
		m.ptyOutput = ""
		output = output[clearIndex+len(clearScreenSequence):]
	}

	m.ptyOutput += readablePTYOutput(output)
	if len(m.ptyOutput) > maxPTYOutput {
		m.ptyOutput = m.ptyOutput[len(m.ptyOutput)-maxPTYOutput:]
	}
}

// readablePTYOutput removes terminal-control sequences because Bubble Tea owns
// the terminal screen and renders the PTY transcript as text.
func readablePTYOutput(output string) string {
	output = ansiOSC.ReplaceAllString(output, "")
	output = ansiCSI.ReplaceAllString(output, "")
	output = crlf.ReplaceAllString(output, "\n")
	return strings.ReplaceAll(output, "\r", "")
}

func (m *Model) handlePTYClosed(message ptyClosedMsg) {
	if m.pty != message.session {
		return
	}

	m.pty.Close()
	m.pty = nil
	m.ptyOutput = ""
	m.screen = menuScreen
}

func (m Model) handlePTYKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pty == nil {
		if message.Type == tea.KeyEnter {
			m.screen = menuScreen
			m.ptyError = nil
		}
		if message.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	}

	if input := ptyInput(message); len(input) > 0 {
		_, _ = m.pty.Write(input)
	}
	return m, nil
}

func readPTYOutput(session *shellpty.Session) tea.Cmd {
	return func() tea.Msg {
		buffer := make([]byte, 4096)
		count, err := session.Read(buffer)
		if count > 0 {
			return ptyOutputMsg{session: session, output: string(buffer[:count])}
		}
		if err != nil {
			return ptyClosedMsg{session: session}
		}
		return ptyClosedMsg{session: session}
	}
}

func ptyInput(message tea.KeyMsg) []byte {
	switch message.Type {
	case tea.KeyRunes:
		input := []byte(string(message.Runes))
		if message.Alt {
			return append([]byte{'\x1b'}, input...)
		}
		return input
	case tea.KeyEnter:
		return []byte{'\r'}
	case tea.KeyBackspace:
		return []byte{'\x7f'}
	case tea.KeyTab:
		return []byte{'\t'}
	case tea.KeySpace:
		return []byte{' '}
	case tea.KeyCtrlC:
		return []byte{'\x03'}
	case tea.KeyCtrlD:
		return []byte{'\x04'}
	case tea.KeyUp:
		return []byte("\x1b[A")
	case tea.KeyDown:
		return []byte("\x1b[B")
	case tea.KeyRight:
		return []byte("\x1b[C")
	case tea.KeyLeft:
		return []byte("\x1b[D")
	default:
		return nil
	}
}

func ptyView(output string, startError error) string {
	if startError != nil {
		return strings.Join([]string{
			titleStyle.Render("Unable to start Bash"),
			startError.Error(),
			"",
			muted.Render("Press Enter to return. Ctrl+C exits."),
		}, "\n")
	}

	if output == "" {
		return muted.Render("Starting shell...")
	}

	return strings.Join([]string{
		muted.Render("Bare Bash session. Ctrl+C interrupts a command; Ctrl+D returns to the menu."),
		"",
		output,
	}, "\n")
}
