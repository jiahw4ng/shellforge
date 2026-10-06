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
		want     string
	}{
		{name: "collapsed", want: "> F1 show hints"},
		{name: "expanded", expanded: true, want: "v F1 hide hints\nFirst hint."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := ansi.Strip(Disclosure(binding, test.expanded, "show hints", "hide hints", "First hint."))
			if view != test.want {
				t.Fatalf("Disclosure() = %q, want %q", view, test.want)
			}
		})
	}
}
