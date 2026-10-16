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

func TestRenderUsesOneAccentWithoutHeadingBanner(t *testing.T) {
	rendered, err := Render(lessons.Markdown("# Heading\n\n## Subheading"), 40)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if occurrences := strings.Count(rendered, "38;5;33"); occurrences < 2 {
		t.Fatalf("Render() accent foreground occurrences = %d, want both headings: %q", occurrences, rendered)
	}
	if strings.Contains(rendered, "48;5;") {
		t.Fatalf("Render() heading includes a background banner: %q", rendered)
	}
}

func TestRenderHintUsesNeutralTextAndInlineCodeEmphasis(t *testing.T) {
	rendered, err := RenderHint(lessons.Markdown("Try `pwd`."), 40)
	if err != nil {
		t.Fatalf("RenderHint() error = %v", err)
	}
	if !strings.Contains(rendered, "38;5;252") {
		t.Fatalf("RenderHint() does not use neutral text: %q", rendered)
	}
	if !strings.Contains(rendered, "38;5;222") || !strings.Contains(rendered, "48;5;238") {
		t.Fatalf("RenderHint() does not retain inline-code emphasis: %q", rendered)
	}
}
