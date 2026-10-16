package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
)

const disclosureRailWidth = 2

const legitEmpty = "There is nothing to show here."

// Disclosure renders a key-controlled section as a compact summary and an
// expanded body, using emptyBody when no body content is available.
func Disclosure(binding key.Binding, expanded bool, showLabel, hideLabel, body, emptyBody string, width int) string {
	description := showLabel
	headerMark := "╶"
	if expanded {
		description = hideLabel
		headerMark = "┌"
	}

	help := binding.Help()
	header := MutedStyle.Render(headerMark + " " + help.Key + ": " + description)
	if !expanded {
		return header
	}
	if body == "" {
		if emptyBody == "" {
			emptyBody = legitEmpty
		}
		body = emptyBody
	}
	return header + "\n" + disclosureBody(body, width)
}

func disclosureBody(body string, width int) string {
	wrapped := ansi.Wrap(body, max(width-disclosureRailWidth, 1), "")
	lines := strings.Split(wrapped, "\n")
	for i := range lines {
		rail := "│"
		if i == len(lines)-1 {
			rail = "└"
		}
		lines[i] = MutedStyle.Render(rail) + " " + lines[i]
	}
	return strings.Join(lines, "\n")
}
