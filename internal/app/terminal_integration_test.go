package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
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

	consumeOuterTerminalUpdate(t, &model, model.term.session.Init())
	if message := model.term.session.SendInput("printf terminal-ok\r")(); message != nil {
		updated, _ := model.Update(message)
		model = updated.(State)
	}
	consumeOuterTerminalUpdate(t, &model, model.term.session.Init())

	if !strings.Contains(model.term.session.View(), "terminal-ok") {
		t.Fatalf("terminal view does not contain command output: %q", model.term.session.View())
	}
	view := model.View()
	if view.Cursor == nil {
		t.Fatal("terminal view cursor = nil, want visible cursor")
	}
	if !view.Cursor.Blink {
		t.Error("terminal view cursor does not blink")
	}
}

func consumeOuterTerminalUpdate(t *testing.T, model *State, command tea.Cmd) {
	t.Helper()

	result := make(chan any, 1)
	go func() { result <- command() }()

	select {
	case message := <-result:
		updated, _ := model.Update(message)
		*model = updated.(State)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for terminal output")
	}
}
