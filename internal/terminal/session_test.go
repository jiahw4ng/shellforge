package terminal

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

const terminalHelperProcessEnvironment = "SHELLFORGE_TERMINAL_HELPER_PROCESS"

func TestObserveProcessExitDetectsFastExit(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=TestTerminalHelperProcess")
	command.Env = append(os.Environ(), terminalHelperProcessEnvironment+"=1")
	if err := command.Start(); err != nil {
		t.Fatalf("helper process start error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-observeProcessExit(ctx, command):
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fast process exit")
	}
}

func TestTerminalHelperProcess(t *testing.T) {
	if os.Getenv(terminalHelperProcessEnvironment) != "1" {
		return
	}
	os.Exit(0)
}
