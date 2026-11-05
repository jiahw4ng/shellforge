package config

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCursorStylesCoverEveryShapeAndBlinkOnce(t *testing.T) {
	type variant struct {
		shape tea.CursorShape
		blink bool
	}
	seenIDs := make(map[string]bool)
	seenVariants := make(map[variant]bool)
	for _, style := range CursorStyles {
		if seenIDs[style.ID] {
			t.Errorf("duplicate cursor style ID %q", style.ID)
		}
		seenIDs[style.ID] = true
		seenVariants[variant{style.Shape, style.Blink}] = true

		found, ok := CursorStyleByID(style.ID)
		if !ok || found != style {
			t.Errorf("CursorStyleByID(%q) = %#v, %v; want %#v", style.ID, found, ok, style)
		}
	}
	for _, shape := range []tea.CursorShape{tea.CursorBlock, tea.CursorUnderline, tea.CursorBar} {
		for _, blink := range []bool{true, false} {
			if !seenVariants[variant{shape, blink}] {
				t.Errorf("no cursor style for shape %d with blink %v", shape, blink)
			}
		}
	}
	if len(CursorStyles) != 6 {
		t.Errorf("cursor styles = %d, want 6", len(CursorStyles))
	}
}

func TestDefaultCursorStyleIsBlinkingBlock(t *testing.T) {
	if DefaultCursorStyle.Shape != tea.CursorBlock || !DefaultCursorStyle.Blink {
		t.Fatalf("DefaultCursorStyle = %#v, want blinking block", DefaultCursorStyle)
	}
}

func TestCursorStyleByIDRejectsUnknownID(t *testing.T) {
	if style, ok := CursorStyleByID("triangle"); ok {
		t.Fatalf("CursorStyleByID(unknown) = %#v, want not found", style)
	}
}
