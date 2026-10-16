package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"
)

func DimensionWithFallback(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func TrimLeadingBlankLines(text string) string {
	lines := lo.DropWhile(strings.Split(text, "\n"), func(line string) bool {
		return strings.TrimSpace(ansi.Strip(line)) == ""
	})
	return strings.Join(lines, "\n")
}
