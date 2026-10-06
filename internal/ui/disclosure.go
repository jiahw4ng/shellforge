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
	if expanded {
		description = hideLabel
	}

	help := binding.Help()
	header := MutedStyle.Render("┌ " + help.Key + ": " + description)
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
	rail := MutedStyle.Render("│") + " "
	lines := strings.Split(wrapped, "\n")
	for i := range lines {
		lines[i] = rail + lines[i]
	}
	return strings.Join(lines, "\n")
}
