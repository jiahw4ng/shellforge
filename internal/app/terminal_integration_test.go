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

	model := State{nav: navigationState{screen: terminalScreen}, term: terminalState{session: started.session}}
	t.Cleanup(model.Close)

	consumeOuterTerminalUpdate(t, &model, model.term.session.Init())
	if message := model.term.session.SendInput("printf bubbleterm-ok\r")(); message != nil {
		updated, _ := model.Update(message)
		model = updated.(State)
	}
	consumeOuterTerminalUpdate(t, &model, model.term.session.Init())

	if !strings.Contains(model.term.session.View(), "bubbleterm-ok") {
		t.Fatalf("terminal view does not contain command output: %q", model.term.session.View())
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
