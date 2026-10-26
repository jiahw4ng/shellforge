//go:build !windows

package terminal

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/xpty"
)

func TestTerminalEngineRunsCommandThroughPTY(t *testing.T) {
	pseudoTerminal, err := xpty.NewPty(20, 10)
	if err != nil {
		t.Fatalf("xpty.NewPty() error = %v", err)
	}
	command := exec.Command("/bin/sh", "-c", "printf pty-ok")
	preparePTYCommand(command)
	if err := pseudoTerminal.Start(command); err != nil {
		_ = pseudoTerminal.Close()
		t.Fatalf("Pty.Start() error = %v", err)
	}

	engine := newTerminalEngine(20, 10, pseudoTerminal)
	ctx, cancel := context.WithCancel(context.Background())
	exited := observeProcessExit(ctx, command)
	t.Cleanup(func() {
		cancel()
		_ = command.Process.Kill()
		_ = engine.close()
		<-exited
	})

	waitForTerminalCondition(t, func() bool {
		return strings.Contains(engine.view(), "pty-ok")
	}, "real PTY command output")
	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for PTY command exit")
	}
}
