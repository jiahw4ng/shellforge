package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"
)

const (
	disclosureRailWidth     = 2
	collapsedDisclosureMark = "▶"
	expandedDisclosureMark  = "▼"
)

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
	headerMark := collapsedDisclosureMark
	if options.Expanded {
		headerMark = expandedDisclosureMark
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
	lines = lo.Map(lines, func(line string, index int) string {
		rail := "│"
		if index == len(lines)-1 {
			rail = "└"
		}
		return MutedStyle.Render(rail) + " " + line
	})
	return strings.Join(lines, "\n")
}
