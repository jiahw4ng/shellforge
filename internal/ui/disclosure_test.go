package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
)

func TestDisclosure(t *testing.T) {
	binding := key.NewBinding(key.WithHelp("F1", "unused"))

	tests := []struct {
		name     string
		expanded bool
		body     string
		want     string
	}{
		{name: "collapsed", want: "┌ F1: show hints"},
		{name: "expanded", expanded: true, body: "First hint.", want: "┌ F1: hide hints\n│ First\n│ hint."},
		{name: "empty", expanded: true, want: "┌ F1: hide hints\n│ Press\n│ F10."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := ansi.Strip(Disclosure(binding, test.expanded, "show hints", "hide hints", test.body, "Press F10.", 7))
			if view != test.want {
				t.Fatalf("Disclosure() = %q, want %q", view, test.want)
			}
		})
	}
}
