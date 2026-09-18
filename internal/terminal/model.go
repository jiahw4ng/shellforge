package terminal

import (
	"shellforge/internal/container"

	bubbleterm "github.com/taigrr/bubbleterm"
)

// TermSession groups the emulator, its Docker sandbox, and its exit notification.
type TermSession struct {
	emulator *bubbleterm.Model
	sandbox  *container.Container
	exited   <-chan struct{}
}
