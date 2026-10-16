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
		{name: "collapsed", want: "▸ F1 Hints"},
		{name: "expanded", expanded: true, body: "First hint.", want: "▾ F1 Hints\n│ First\n└ hint."},
		{name: "empty", expanded: true, want: "▾ F1 Hints\n│ Press\n└ F10."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := ansi.Strip(Disclosure(DisclosureOptions{
				Binding:   binding,
				Expanded:  test.expanded,
				Label:     "Hints",
				Body:      test.body,
				EmptyBody: "Press F10.",
				Width:     7,
			}))
			if view != test.want {
				t.Fatalf("Disclosure() = %q, want %q", view, test.want)
			}
		})
	}
}
