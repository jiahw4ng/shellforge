package terminal

import (
	"context"
	"os"
	"shellforge/internal/container"
	"sync"
)

// TermSession groups the terminal engine, its Docker sandbox, and its exit notification.
type TermSession struct {
	terminal      *terminalEngine
	sandbox       *container.Container
	cancelProcess context.CancelFunc
	process       *os.Process
	closeOnce     sync.Once
	Exited        <-chan struct{}
}
