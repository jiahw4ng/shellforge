package app

import (
	"errors"
	"shellforge/internal/config"
	"shellforge/internal/container"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestTerminalRunsCommandInLessonContainer verifies command output travels
// through Docker, the terminal engine, and Shellforge's outer event loop.
func TestTerminalRunsCommandInLessonContainer(t *testing.T) {
	started := awaitTerminalStarted(t, launchTerminal(80, 24, nil, 0))
	if errors.Is(started.err, container.ErrContainerUnavailable) {
		t.Skip(started.err)
	}
	if started.err != nil {
		t.Fatalf("launchTerminal() error = %v", started.err)
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

// startIntegrationSandbox starts a real standalone sandbox for a test, skipping
// when Docker is unavailable, and waits for the learner prompt.
func startIntegrationSandbox(t *testing.T) State {
	t.Helper()
	started := awaitTerminalStarted(t, launchTerminal(80, 24, nil, 0))
	if errors.Is(started.err, container.ErrContainerUnavailable) {
		t.Skip(started.err)
	}
	if started.err != nil {
		t.Fatalf("launchTerminal() error = %v", started.err)
	}
	model := State{
		nav:      navigationState{screen: config.SandboxTerminalScreen},
		term:     terminalState{session: started.session},
		viewport: viewportState{width: 84, height: 27},
	}
	t.Cleanup(started.session.Close)
	waitForTerminalLine(t, &model, "student:~/workspace$")
	return model
}

// typeIntoTerminal sends raw input to the shell through the outer event loop.
func typeIntoTerminal(model *State, input string) {
	if message := model.term.session.SendInput(input)(); message != nil {
		updated, _ := model.Update(message)
		*model = updated.(State)
	}
}

// awaitTerminalExit runs the session's exit command, including its check of
// the sandbox, and fails the test if the shell has not exited in ten seconds.
func awaitTerminalExit(t *testing.T, model State) termExitedMsg {
	t.Helper()
	result := make(chan tea.Msg, 1)
	go func() { result <- waitForTerminalExit(model.term.session, model.term.gen)() }()
	select {
	case message := <-result:
		return message.(termExitedMsg)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the shell to exit")
		return termExitedMsg{}
	}
}

// TestOnlyExitLeavesTheShell checks that Ctrl+D on an empty prompt only gets
// Bash's hint, and that typing exit returns to the menu without a failure.
func TestOnlyExitLeavesTheShell(t *testing.T) {
	model := startIntegrationSandbox(t)

	typeIntoTerminal(&model, "\x04") // Ctrl+D on an empty prompt
	waitForTerminalLine(t, &model, `Use "exit" to leave the shell.`)

	typeIntoTerminal(&model, "exit\r")
	exited := awaitTerminalExit(t, model)
	if !exited.sandboxRunning {
		t.Fatal("sandboxRunning = false after typing exit, want the sandbox still running")
	}
	updated, _ := model.Update(exited)
	model = updated.(State)
	if model.nav.screen != config.MenuScreen || model.term.err != nil || model.term.session != nil {
		t.Errorf("after exit: screen = %d, err = %v, session = %v; want the menu with no failure and no session",
			model.nav.screen, model.term.err, model.term.session)
	}
}

// TestRemovedSandboxShowsFailure checks that a shell ending because its
// sandbox was removed is reported as a failure, not as leaving the shell.
func TestRemovedSandboxShowsFailure(t *testing.T) {
	model := startIntegrationSandbox(t)

	// Close kills the shell and force-removes the container, as deleting it
	// from another window would.
	model.term.session.Close()
	exited := awaitTerminalExit(t, model)
	if exited.sandboxRunning {
		t.Fatal("sandboxRunning = true after the sandbox was removed")
	}
	updated, _ := model.Update(exited)
	model = updated.(State)
	if model.nav.screen != config.SandboxTerminalScreen || model.term.err == nil {
		t.Errorf("after removal: screen = %d, err = %v; want the sandbox screen showing a failure",
			model.nav.screen, model.term.err)
	}
}

// awaitTerminalStarted reads a start attempt's events, skipping progress
// messages, until the final result arrives.
func awaitTerminalStarted(t *testing.T, events <-chan tea.Msg) termStartedMsg {
	t.Helper()
	for message := range events {
		if started, ok := message.(termStartedMsg); ok {
			return started
		}
	}
	t.Fatal("start events closed without a termStartedMsg")
	return termStartedMsg{}
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
