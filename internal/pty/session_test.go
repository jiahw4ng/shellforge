package pty

import (
	"strings"
	"testing"
	"time"
)

func TestSessionRunsBashAndReturnsOutput(t *testing.T) {
	session, err := Start(80, 24)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(session.Close)

	if _, err := session.Write([]byte("printf %s pty-; printf %s test\r")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	readUntil(t, session, "pty-test")
}

func TestSessionDisablesNano(t *testing.T) {
	session, err := Start(80, 24)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(session.Close)

	if _, err := session.Write([]byte("nano\r")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	readUntil(t, session, "nano is disabled in Shellforge.")
}

func readUntil(t *testing.T, session *Session, expected string) {
	t.Helper()

	deadline := time.After(3 * time.Second)
	output := ""
	for !strings.Contains(output, expected) {
		result := make(chan readResult, 1)
		go func() {
			buffer := make([]byte, 1024)
			count, readErr := session.Read(buffer)
			result <- readResult{output: string(buffer[:count]), err: readErr}
		}()

		select {
		case read := <-result:
			if read.err != nil {
				t.Fatalf("Read() error = %v", read.err)
			}
			output += read.output
		case <-deadline:
			t.Fatalf("did not receive %q; received %q", expected, output)
		}
	}
}

type readResult struct {
	output string
	err    error
}
