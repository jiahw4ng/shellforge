package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// TestTerminalRunsCommandInLessonContainer verifies command output travels
// through Docker, the terminal engine, and Shellforge's outer event loop.
func TestTerminalRunsCommandInLessonContainer(t *testing.T) {
	message := startTerminal(80, 24, nil, 0)()
	started, ok := message.(termStartedMsg)
	if !ok {
		t.Fatalf("startTerminal() returned %T, want TerminalStartedMsg", message)
	}
	if errors.Is(started.err, container.ErrContainerUnavailable) {
		t.Skip(started.err)
	}
	if started.err != nil {
		t.Fatalf("startTerminal() error = %v", started.err)
	}

	model := State{
		nav:      navigationState{screen: config.SandboxTerminalScreen},
		term:     terminalState{session: started.session},
		viewport: viewportState{width: 84, height: 27},
	}
	t.Cleanup(model.CloseTerminal)

	// Output arrives in several chunks (the first start-up notifications carry
	// nothing printable), so wait for content instead of counting notifications.
	waitForTerminalLine(t, &model, "student:~/workspace$")
	if message := model.term.session.SendInput("printf terminal-ok\r")(); message != nil {
		updated, _ := model.Update(message)
		model = updated.(State)
	}
	// The typed command echoes as "...$ printf terminal-ok", so only the
	// command's own output puts "terminal-ok" at the start of a line.
	waitForTerminalLine(t, &model, "terminal-ok")
	view := model.View()
	if view.Cursor == nil {
		t.Fatal("terminal view cursor = nil, want visible cursor")
	}
	if !view.Cursor.Blink {
		t.Error("terminal view cursor does not blink")
	}
}

// waitForTerminalLine feeds the session's change notifications through the
// outer event loop, as the running application does, until a rendered line
// starts with prefix. It fails the test if that takes longer than five seconds.
func waitForTerminalLine(t *testing.T, model *State, prefix string) {
	t.Helper()

	deadline := time.After(5 * time.Second)
	for !terminalHasLine(model.term.session.View(), prefix) {
		command := model.term.session.Init()
		result := make(chan any, 1)
		go func() { result <- command() }()

		select {
		case message := <-result:
			updated, _ := model.Update(message)
			*model = updated.(State)
		case <-deadline:
			t.Fatalf("timed out waiting for a terminal line starting with %q; view: %q",
				prefix, ansi.Strip(model.term.session.View()))
		}
	}
}

func terminalHasLine(view, prefix string) bool {
	for _, line := range strings.Split(ansi.Strip(view), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}
