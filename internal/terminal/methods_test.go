package terminal

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

type testTerminalTransport struct {
	outputReader *io.PipeReader
	outputWriter *io.PipeWriter
	mu           sync.Mutex
	input        bytes.Buffer
	width        int
	height       int
	closeCount   int
	closeOnce    sync.Once
}

func newTestTerminalTransport() *testTerminalTransport {
	reader, writer := io.Pipe()
	return &testTerminalTransport{outputReader: reader, outputWriter: writer}
}

func (t *testTerminalTransport) Read(buffer []byte) (int, error) {
	return t.outputReader.Read(buffer)
}

func (t *testTerminalTransport) Write(buffer []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.input.Write(buffer)
}

func (t *testTerminalTransport) Resize(width, height int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.width = width
	t.height = height
	return nil
}

func (t *testTerminalTransport) Close() error {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		t.closeCount++
		t.mu.Unlock()
		_ = t.outputWriter.Close()
		_ = t.outputReader.Close()
	})
	return nil
}

func (t *testTerminalTransport) sendOutput(output string) error {
	_, err := io.WriteString(t.outputWriter, output)
	return err
}

func (t *testTerminalTransport) receivedInput() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.input.String()
}

func newTestSession(t *testing.T, width, height int) (*TermSession, *testTerminalTransport) {
	t.Helper()
	transport := newTestTerminalTransport()
	engine := newTerminalEngine(width, height, transport)
	t.Cleanup(func() {
		if err := engine.close(); err != nil {
			t.Errorf("terminalEngine.close() error = %v", err)
		}
	})
	return &TermSession{terminal: engine}, transport
}

func waitForTerminalCondition(t *testing.T, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", description)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestViewRendersTerminalOutput(t *testing.T) {
	session, transport := newTestSession(t, 20, 10)
	if err := transport.sendOutput("shellforge$ printf '\\e[32mok\\e[0m'\r\nok"); err != nil {
		t.Fatalf("sendOutput() error = %v", err)
	}

	waitForTerminalCondition(t, func() bool {
		return strings.Contains(session.View(), "shellforge$") && strings.Contains(session.View(), "ok")
	}, "rendered terminal output")
}

func TestCursorReturnsBlinkingUnderlineAtEmulatorPosition(t *testing.T) {
	session, transport := newTestSession(t, 20, 10)
	if err := transport.sendOutput("\x1b[3;4H"); err != nil {
		t.Fatalf("sendOutput() error = %v", err)
	}

	waitForTerminalCondition(t, func() bool {
		cursor := session.Cursor()
		return cursor != nil && cursor.X == 3 && cursor.Y == 2
	}, "cursor position (3,2)")
	cursor := session.Cursor()
	if !cursor.Blink {
		t.Error("Cursor().Blink = false, want a blinking cursor")
	}
	if cursor.Shape != tea.CursorUnderline {
		t.Errorf("Cursor().Shape = %d, want underline cursor", cursor.Shape)
	}
}

func TestCursorReturnsNilWhenEmulatorHidesCursor(t *testing.T) {
	session, transport := newTestSession(t, 20, 10)
	if err := transport.sendOutput("\x1b[?25l"); err != nil {
		t.Fatalf("sendOutput() error = %v", err)
	}

	waitForTerminalCondition(t, func() bool {
		return session.Cursor() == nil
	}, "hidden cursor")
}

func TestUpdateEncodesKeyThroughVirtualTerminal(t *testing.T) {
	session, transport := newTestSession(t, 20, 10)
	command := session.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if message := command(); message != nil {
		t.Fatalf("Update command message = %#v, want nil", message)
	}

	waitForTerminalCondition(t, func() bool {
		return transport.receivedInput() == "\x1b[A"
	}, "encoded up-arrow input")
}

func TestSendInputAndResizeReachTerminalTransport(t *testing.T) {
	session, transport := newTestSession(t, 20, 10)
	if message := session.SendInput("printf test\r")(); message != nil {
		t.Fatalf("SendInput command message = %#v, want nil", message)
	}
	if message := session.Resize(40, 15)(); message != nil {
		t.Fatalf("Resize command message = %#v, want nil", message)
	}

	waitForTerminalCondition(t, func() bool {
		transport.mu.Lock()
		defer transport.mu.Unlock()
		return transport.input.String() == "printf test\r" && transport.width == 40 && transport.height == 15
	}, "terminal input and resize")
}

func TestUpdateIgnoresAnotherTerminalsOutputMessage(t *testing.T) {
	session, _ := newTestSession(t, 20, 10)
	otherSession, _ := newTestSession(t, 20, 10)

	if command := session.Update(terminalChangedMsg{terminal: otherSession.terminal}); command != nil {
		t.Fatal("Update() returned a command for another terminal's output")
	}
}

func TestTerminalEngineCloseIsIdempotent(t *testing.T) {
	transport := newTestTerminalTransport()
	engine := newTerminalEngine(20, 10, transport)
	if err := engine.close(); err != nil {
		t.Fatalf("first close() error = %v", err)
	}
	if err := engine.close(); err != nil {
		t.Fatalf("second close() error = %v", err)
	}

	transport.mu.Lock()
	defer transport.mu.Unlock()
	if transport.closeCount != 1 {
		t.Fatalf("transport close count = %d, want 1", transport.closeCount)
	}
}
