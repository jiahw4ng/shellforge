package lessonrender

import (
	"shellforge/internal/lessons"
	"strings"
	"testing"
)

func TestRenderStylesInlineAndFencedCode(t *testing.T) {
	rendered, err := Render(lessons.Markdown("Run `pwd`.\n\n```bash\npwd\n```"), 40)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, text := range []string{"Run", "pwd", "\x1b["} {
		if !strings.Contains(rendered, text) {
			t.Errorf("rendered Markdown does not contain %q: %q", text, rendered)
		}
	}
}

func TestRenderCachesByContentAndWidth(t *testing.T) {
	first, err := Render(lessons.Markdown("`pwd`"), 20)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, err := Render(lessons.Markdown("`pwd`"), 20)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}
	if first != second {
		t.Fatalf("cached render differs: first %q, second %q", first, second)
	}
}
