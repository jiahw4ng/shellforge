package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWithAppFrameJunctionJoinsVerticalDivider(t *testing.T) {
	view := ansi.Strip(WithAppFrameJunction("left│right\n    │", 16, 4, 4))
	lines := strings.Split(view, "\n")
	if len(lines) != 4 {
		t.Fatalf("frame height = %d, want 4: %q", len(lines), view)
	}

	junctionColumn := ApplicationBorderSize/2 + ApplicationHorizontalMargin + 4
	for row, want := range map[int]rune{0: '┬', 1: '│', len(lines) - 1: '┴'} {
		cells := []rune(lines[row])
		if got := cells[junctionColumn]; got != want {
			t.Errorf("frame row %d junction = %q, want %q: %q", row, got, want, lines[row])
		}
	}
}

func TestWithAppFrameJunctionIgnoresOutsideContent(t *testing.T) {
	got := WithAppFrameJunction("content", 16, 4, 12)
	want := WithAppFrame("content", 16, 4)
	if got != want {
		t.Fatalf("out-of-range junction changed frame:\ngot  %q\nwant %q", got, want)
	}
}
