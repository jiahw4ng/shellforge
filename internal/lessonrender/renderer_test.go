package lessonrender

import (
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
	"testing"
)

func TestRenderFollowsActiveThemeWithoutStaleCache(t *testing.T) {
	t.Cleanup(func() { ui.ApplyTheme(ui.DefaultTheme) })
	markdown := lessons.Markdown("# Heading\n\nRun `pwd`.")

	renderIn := func(themeID string) string {
		t.Helper()
		theme, ok := ui.ThemeByID(themeID)
		if !ok {
			t.Fatalf("unknown theme %q", themeID)
		}
		ui.ApplyTheme(theme)
		rendered, err := Render(markdown, 40)
		if err != nil {
			t.Fatalf("Render() in %s error = %v", themeID, err)
		}
		return rendered
	}

	dark := renderIn("dark")
	light := renderIn("light")
	highContrast := renderIn("high-contrast")

	if dark == light || dark == highContrast || light == highContrast {
		t.Fatalf("themes rendered identically:\ndark: %q\nlight: %q\nhigh-contrast: %q", dark, light, highContrast)
	}
	for _, test := range []struct{ name, rendered, accent string }{
		{"dark", dark, "38;5;33"},
		{"light", light, "38;5;25"},
		{"high-contrast", highContrast, "38;5;51"},
	} {
		if !strings.Contains(test.rendered, test.accent) {
			t.Errorf("%s render lacks accent %s: %q", test.name, test.accent, test.rendered)
		}
	}
	if again := renderIn("dark"); again != dark {
		t.Fatalf("dark render changed after visiting other themes: %q vs %q", again, dark)
	}
}

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
