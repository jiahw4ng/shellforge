package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

// Disclosure renders a key-controlled section as a compact summary and an
// optional expanded body.
func Disclosure(binding key.Binding, expanded bool, showLabel, hideLabel, body string) string {
	marker := ">"
	description := showLabel
	if expanded {
		marker = "v"
		description = hideLabel
	}

	help := binding.Help()
	header := MutedStyle.Render(strings.Join([]string{marker, help.Key, description}, " "))
	if !expanded || body == "" {
		return header
	}
	return header + "\n" + body
}
