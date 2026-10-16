// Package lessonrender turns lesson Markdown into Shellforge-styled ANSI text.
package lessonrender

import (
	"shellforge/internal/lessons"
	"shellforge/internal/ui"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

type cacheKey struct {
	markdown lessons.Markdown
	width    int
	variant  renderVariant
}

type renderVariant uint8

const (
	lessonVariant renderVariant = iota
	hintVariant
)

const (
	inlineCodeColorCode      = "222"
	inlineCodeBackgroundCode = "238"
)

var renderedPages sync.Map

// Render converts Markdown into styled terminal text wrapped to width. Lesson
// pages are immutable after loading, so caching avoids reparsing Markdown on
// every terminal keypress and redraw.
func Render(markdown lessons.Markdown, width int) (string, error) {
	return render(markdown, width, lessonVariant)
}

// RenderHint converts hint Markdown into neutral instructional text.
func RenderHint(markdown lessons.Markdown, width int) (string, error) {
	return render(markdown, width, hintVariant)
}

func render(markdown lessons.Markdown, width int, variant renderVariant) (string, error) {
	if width < 1 {
		width = 1
	}

	key := cacheKey{markdown: markdown, width: width, variant: variant}
	if cached, ok := renderedPages.Load(key); ok {
		return cached.(string), nil
	}

	style := shellforgeStyle()
	if variant == hintVariant {
		style = hintStyle()
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return "", err
	}

	rendered, err := renderer.Render(string(markdown))
	if err != nil {
		return "", err
	}
	rendered = strings.TrimSpace(rendered)
	renderedPages.Store(key, rendered)
	return rendered, nil
}

func hintStyle() ansi.StyleConfig {
	return shellforgeStyle()
}

// shellforgeStyle starts from Glamour's dark terminal palette and narrows it
// for instructional content: commands are conspicuous without making prose
// noisy, while fenced blocks retain Bash syntax highlighting.
func shellforgeStyle() ansi.StyleConfig {
	style := styles.DarkStyleConfig
	style.Document.Margin = new(uint(0))
	style.Document.StylePrimitive.Color = new(ui.TextColorCode)
	style.Heading.StylePrimitive.Color = new(ui.AccentColorCode)
	style.Heading.StylePrimitive.Bold = new(true)
	style.H1.StylePrimitive.Prefix = ""
	style.H1.StylePrimitive.Suffix = ""
	style.H1.StylePrimitive.Color = new(ui.AccentColorCode)
	style.H1.StylePrimitive.BackgroundColor = nil
	style.H2.StylePrimitive.Color = new(ui.AccentColorCode)
	style.H2.StylePrimitive.Prefix = ""
	style.H2.StylePrimitive.BlockPrefix = ""
	style.H2.StylePrimitive.BlockSuffix = "\n"
	style.H3.StylePrimitive.Color = new(ui.AccentColorCode)
	style.H4.StylePrimitive.Color = new(ui.AccentColorCode)
	style.H5.StylePrimitive.Color = new(ui.AccentColorCode)
	style.H6.StylePrimitive.Color = new(ui.AccentColorCode)
	style.Code.StylePrimitive.Color = new(inlineCodeColorCode)
	style.Code.StylePrimitive.BackgroundColor = new(inlineCodeBackgroundCode)
	style.Code.StylePrimitive.Bold = new(true)
	style.CodeBlock.StyleBlock.Margin = new(uint(0))
	return style
}
