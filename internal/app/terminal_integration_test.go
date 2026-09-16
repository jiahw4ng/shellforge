package app

import (
	"errors"
	"shellforge/internal/container"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestTerminalRunsCommandInLessonContainer verifies command output travels
// through Docker, Bubbleterm, and Shellforge's outer event loop.
func TestTerminalRunsCommandInLessonContainer(t *testing.T) {
	message := startTerminal(80, 24)()
	started, ok := message.(TerminalStartedMsg)
	if !ok {
		t.Fatalf("startTerminal() returned %T, want TerminalStartedMsg", message)
	}
	if errors.Is(started.Err, container.ErrUnavailable) {
		t.Skip(started.Err)
	}
	if started.Err != nil {
		t.Fatalf("startTerminal() error = %v", started.Err)
	}

	model := State{CurrentScreen: terminalScreen, Terminal: started.Session}
	t.Cleanup(model.Close)

	consumeOuterTerminalUpdate(t, &model, model.Terminal.Init())
	if message := model.Terminal.SendInput("printf bubbleterm-ok\r")(); message != nil {
		updated, _ := model.Update(message)
		model = updated.(State)
	}
	consumeOuterTerminalUpdate(t, &model, model.Terminal.Init())

	if !strings.Contains(model.Terminal.View(), "bubbleterm-ok") {
		t.Fatalf("terminal view does not contain command output: %q", model.Terminal.View())
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
