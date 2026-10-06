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
	variant  renderVariant
}

type renderVariant uint8

const (
	lessonVariant renderVariant = iota
	hintVariant
)

var renderedPages sync.Map

// Render converts Markdown into styled terminal text wrapped to width. Lesson
// pages are immutable after loading, so caching avoids reparsing Markdown on
// every terminal keypress and redraw.
func Render(markdown lessons.Markdown, width int) (string, error) {
	return render(markdown, width, lessonVariant)
}

// RenderHint converts hint Markdown into yellow styled terminal text.
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
	style := shellforgeStyle()
	yellow := "226"
	style.Document.StylePrimitive.Color = &yellow
	style.BlockQuote.StylePrimitive.Color = &yellow
	style.Paragraph.StylePrimitive.Color = &yellow
	style.List.StylePrimitive.Color = &yellow
	style.Heading.StylePrimitive.Color = &yellow
	style.H1.StylePrimitive.Color = &yellow
	style.H2.StylePrimitive.Color = &yellow
	style.H3.StylePrimitive.Color = &yellow
	style.H4.StylePrimitive.Color = &yellow
	style.H5.StylePrimitive.Color = &yellow
	style.H6.StylePrimitive.Color = &yellow
	style.Text.Color = &yellow
	style.Strikethrough.Color = &yellow
	style.Emph.Color = &yellow
	style.Strong.Color = &yellow
	style.HorizontalRule.Color = &yellow
	style.Item.Color = &yellow
	style.Enumeration.Color = &yellow
	style.Task.StylePrimitive.Color = &yellow
	style.Link.Color = &yellow
	style.LinkText.Color = &yellow
	style.Image.Color = &yellow
	style.ImageText.Color = &yellow
	style.Code.StylePrimitive.Color = &yellow
	style.CodeBlock.StylePrimitive.Color = &yellow
	style.Table.StylePrimitive.Color = &yellow
	style.DefinitionList.StylePrimitive.Color = &yellow
	style.DefinitionTerm.Color = &yellow
	style.DefinitionDescription.Color = &yellow
	style.HTMLBlock.StylePrimitive.Color = &yellow
	style.HTMLSpan.StylePrimitive.Color = &yellow
	return style
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
