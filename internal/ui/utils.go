package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func DimensionWithFallback(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func TrimLeadingBlankLines(text string) string {
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && strings.TrimSpace(ansi.Strip(lines[0])) == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}
