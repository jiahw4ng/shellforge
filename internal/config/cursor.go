package config

import tea "charm.land/bubbletea/v2"

// CursorStyle is one learner-selectable shape and blink combination for the
// terminal cursor. ID is the stable value persisted in learner settings.
type CursorStyle struct {
	ID    string
	Label string
	Shape tea.CursorShape
	Blink bool
}

// CursorStyles lists every cursor shape and blink combination Bubble Tea supports.
var CursorStyles = []CursorStyle{
	{ID: "block-blinking", Label: "Block (blinking)", Shape: tea.CursorBlock, Blink: true},
	{ID: "block-steady", Label: "Block (steady)", Shape: tea.CursorBlock, Blink: false},
	{ID: "underline-blinking", Label: "Underline (blinking)", Shape: tea.CursorUnderline, Blink: true},
	{ID: "underline-steady", Label: "Underline (steady)", Shape: tea.CursorUnderline, Blink: false},
	{ID: "bar-blinking", Label: "Bar (blinking)", Shape: tea.CursorBar, Blink: true},
	{ID: "bar-steady", Label: "Bar (steady)", Shape: tea.CursorBar, Blink: false},
}

// DefaultCursorStyle is used until the learner chooses or loads another style.
var DefaultCursorStyle = CursorStyles[0]

// CursorStyleByID returns the cursor style with the given persisted ID.
func CursorStyleByID(id string) (CursorStyle, bool) {
	for _, style := range CursorStyles {
		if style.ID == id {
			return style, true
		}
	}
	return CursorStyle{}, false
}
