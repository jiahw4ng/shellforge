package terminal

import (
	"strings"
	"testing"
)

func TestRenderCursorHighlightsTheCursorCell(t *testing.T) {
	view := renderCursor([]string{"shellforge$ "}, 11, 0)
	if !strings.Contains(view, "shellforge$\x1b[7m \x1b[0m") {
		t.Fatalf("cursor is not highlighted in %q", view)
	}
}

func TestRenderCursorLeavesInvalidPositionUnchanged(t *testing.T) {
	view := renderCursor([]string{"shellforge$ "}, 99, 0)
	if view != "shellforge$ " {
		t.Fatalf("invalid cursor position changed view to %q", view)
	}
}
