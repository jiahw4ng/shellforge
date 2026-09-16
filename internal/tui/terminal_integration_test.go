package tui

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
	started, ok := message.(terminalStartedMsg)
	if !ok {
		t.Fatalf("startTerminal() returned %T, want terminalStartedMsg", message)
	}
	if errors.Is(started.err, container.ErrUnavailable) {
		t.Skip(started.err)
	}
	if started.err != nil {
		t.Fatalf("startTerminal() error = %v", started.err)
	}

	model := Model{
		currentScreen:   terminalScreen,
		terminal:        started.terminal,
		lessonContainer: started.lessonContainer,
		terminalExit:    started.exited,
	}
	t.Cleanup(model.Close)

	consumeOuterTerminalUpdate(t, &model, model.terminal.Init())
	if message := model.terminal.SendInput("printf bubbleterm-ok\\r")(); message != nil {
		updated, _ := model.Update(message)
		model = updated.(Model)
	}
	consumeOuterTerminalUpdate(t, &model, model.terminal.Init())

	if !strings.Contains(model.terminal.View().Content, "bubbleterm-ok") {
		t.Fatalf("terminal view does not contain command output: %q", model.terminal.View().Content)
	}
}

// consumeOuterTerminalUpdate waits for one Bubbleterm event and feeds it back
// through Shellforge, matching the real Bubble Tea program loop.
func consumeOuterTerminalUpdate(t *testing.T, model *Model, command tea.Cmd) {
	t.Helper()

	result := make(chan any, 1)
	go func() { result <- command() }()

	select {
	case message := <-result:
		updated, _ := model.Update(message)
		*model = updated.(Model)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for terminal output")
	}
}
