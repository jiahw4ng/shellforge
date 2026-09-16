package terminal

import (
	"shellforge/internal/container"

	bubbleterm "github.com/taigrr/bubbleterm"
)

// Session groups the emulator, its Docker sandbox, and its exit notification.
type Session struct {
	emulator *bubbleterm.Model
	sandbox  *container.Container
	exited   <-chan struct{}
}
