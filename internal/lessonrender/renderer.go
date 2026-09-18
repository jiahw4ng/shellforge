// Package lessonrender turns lesson Markdown into Shellforge-styled ANSI text.
package lessonrender

import (
	"shellforge/internal/lessons"
	"strings"
	"sync"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
)

type cacheKey struct {
	markdown lessons.Markdown
	width    int
}

var renderedPages sync.Map

// Render converts Markdown into styled terminal text wrapped to width. Lesson
// pages are immutable after loading, so caching avoids reparsing Markdown on
// every terminal keypress and redraw.
func Render(markdown lessons.Markdown, width int) (string, error) {
	if width < 1 {
		width = 1
	}

	key := cacheKey{markdown: markdown, width: width}
	if cached, ok := renderedPages.Load(key); ok {
		return cached.(string), nil
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(shellforgeStyle()),
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

// shellforgeStyle starts from Glamour's dark terminal palette and narrows it
// for instructional content: commands are conspicuous without making prose
// noisy, while fenced blocks retain Bash syntax highlighting.
func shellforgeStyle() ansi.StyleConfig {
	style := styles.DarkStyleConfig
	style.Document.Margin = new(uint(0))
	style.Heading.StylePrimitive.Color = new("81")
	style.Heading.StylePrimitive.Bold = new(true)
	style.H2.StylePrimitive.Prefix = ""
	style.H2.StylePrimitive.BlockPrefix = ""
	style.H2.StylePrimitive.BlockSuffix = "\n"
	style.Code.StylePrimitive.Color = new("222")
	style.Code.StylePrimitive.BackgroundColor = new("238")
	style.Code.StylePrimitive.Bold = new(true)
	style.CodeBlock.StyleBlock.Margin = new(uint(0))
	return style
}
