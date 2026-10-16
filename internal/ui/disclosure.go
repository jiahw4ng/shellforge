package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
)

const disclosureRailWidth = 2

const legitEmpty = "There is nothing to show here."

// DisclosureOptions contains the binding, label, content, and layout state for
// one disclosure panel.
type DisclosureOptions struct {
	Binding   key.Binding
	Expanded  bool
	Label     string
	Body      string
	EmptyBody string
	Width     int
}

// Disclosure renders a key-controlled section as a compact summary and an
// expanded body, using EmptyBody when no body content is available.
func Disclosure(options DisclosureOptions) string {
	headerMark := "▸"
	if options.Expanded {
		headerMark = "▾"
	}

	help := options.Binding.Help()
	header := ControlStyle.Render(headerMark + " " + help.Key + " " + options.Label)
	if !options.Expanded {
		return header
	}
	if options.Body == "" {
		if options.EmptyBody == "" {
			options.EmptyBody = legitEmpty
		}
		options.Body = options.EmptyBody
	}
	return header + "\n" + disclosureBody(options.Body, options.Width)
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
