package terminal

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	bubbleterm "github.com/taigrr/bubbleterm"
)

type discardWriteCloser struct {
	io.Writer
}

func (discardWriteCloser) Close() error {
	return nil
}

func TestCursorReturnsBlinkingBlockAtEmulatorPosition(t *testing.T) {
	emulator, err := bubbleterm.NewWithPipes(
		20,
		10,
		strings.NewReader("\x1b[3;4H"),
		discardWriteCloser{Writer: io.Discard},
	)
	if err != nil {
		t.Fatalf("NewWithPipes() error = %v", err)
	}
	t.Cleanup(func() { _ = emulator.Close() })
	session := TermSession{emulator: emulator}

	deadline := time.Now().Add(time.Second)
	for {
		cursor := session.Cursor()
		if cursor != nil && cursor.X == 3 && cursor.Y == 2 {
			if !cursor.Blink {
				t.Error("Cursor().Blink = false, want a blinking cursor")
			}
			if cursor.Shape != tea.CursorUnderline {
				t.Errorf("Cursor().Shape = %d, want underline cursor", cursor.Shape)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Cursor() did not reach emulator position (3,2): %#v", cursor)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestCursorReturnsNilWhenEmulatorHidesCursor(t *testing.T) {
	emulator, err := bubbleterm.NewWithPipes(
		20,
		10,
		strings.NewReader("\x1b[?25l"),
		discardWriteCloser{Writer: io.Discard},
	)
	if err != nil {
		t.Fatalf("NewWithPipes() error = %v", err)
	}
	t.Cleanup(func() { _ = emulator.Close() })
	session := TermSession{emulator: emulator}

	deadline := time.Now().Add(time.Second)
	for {
		_, visible := emulator.GetEmulator().Cursor()
		if !visible {
			if cursor := session.Cursor(); cursor != nil {
				t.Fatalf("Cursor() = %#v, want nil for a hidden emulator cursor", cursor)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("emulator cursor remained visible")
		}
		time.Sleep(time.Millisecond)
	}
}
