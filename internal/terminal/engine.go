package terminal

import (
	"errors"
	"io"
	"log/slog"
	"shellforge/internal/ui"
	"sync"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
)

// terminalTransport is the host side of the pseudo-terminal: reading returns
// what the shell inside the container writes, and writing delivers keystrokes to
// it. In production it is the xpty PTY or ConPTY running `docker exec -it`;
// tests substitute an in-memory pipe.
type terminalTransport interface {
	io.ReadWriteCloser
	Resize(width, height int) error
}

// terminalChangedMsg tells Bubble Tea that the screen may have changed and
// should be redrawn. terminal identifies the engine, so a message from a
// replaced or reset session is ignored.
type terminalChangedMsg struct {
	terminal *terminalEngine
}

// terminalEngine connects a shell process to a virtual terminal emulator. Two
// goroutines move bytes between them:
//
//	shell --(transport.Read)--> readProcessOutput --(emulator.Write)--> screen grid
//	keys  --(emulator.SendKey)--> writeTerminalResponses --(transport.Write)--> shell
//
// Bubble Tea never touches the transport. It reads the screen through view and
// cursor, and sends keys through sendKey and sendText.
type terminalEngine struct {
	// mu guards emulator and the render cache. The emulator is not safe for
	// concurrent use, and the output goroutine, key input, and rendering all
	// reach it from different goroutines.
	mu sync.RWMutex
	// emulator parses ANSI output into a screen grid and encodes key input.
	emulator *vt.Emulator
	// transport is the PTY attached to the shell process.
	transport terminalTransport
	// changed holds at most one pending "screen changed" signal, so bursts of
	// output coalesce into one redraw instead of queueing one per chunk.
	changed chan struct{}
	// done is closed once by close; both goroutines and waitForChange watch it.
	done chan struct{}
	// cursorVisible follows the shell's show/hide-cursor sequences.
	cursorVisible bool
	closeOnce     sync.Once
	// workers waits for both I/O goroutines to exit during close.
	workers sync.WaitGroup

	// version counts screen changes (output and resize). It is guarded by mu.
	version uint64
	// frame and the cached* fields hold the last themed render, so frames that
	// follow no screen or theme change reuse it. They are guarded by mu.
	frame         *uv.Buffer
	cachedView    string
	cachedVersion uint64
	cachedText    [4]uint32
	hasCachedView bool
}

// newTerminalEngine starts the two I/O goroutines over transport. The caller
// must call close to stop them and release the transport.
func newTerminalEngine(width, height int, transport terminalTransport) *terminalEngine {
	engine := &terminalEngine{
		// the virtual "screen" that the terminal renders to
		emulator: vt.NewEmulator(width, height),
		// communication with Docker
		transport: transport,
		// screen update notifications
		changed: make(chan struct{}, 1),
		// shutdown
		done:          make(chan struct{}),
		cursorVisible: true,
	}
	engine.emulator.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			engine.cursorVisible = visible
		},
	})
	engine.workers.Add(2)
	// handle io streams
	go engine.readProcessOutput()
	go engine.writeTerminalResponses()
	return engine
}

// readProcessOutput is the output pump: everything the shell prints reaches the
// screen through this loop. It runs on its own goroutine for the engine's
// whole life.
//
// Each iteration blocks in transport.Read until the shell writes something
// (at most 4 KiB at a time), then feeds the bytes to the emulator. The emulator
// parses text, colors, cursor movement, and scrolling into its screen grid.
// Parsing happens under mu, so key input and rendering wait for the current
// chunk to finish. Bumping version marks the cached render stale, and
// notifyChanged asks Bubble Tea for a redraw without blocking this loop.
//
// The emulator may answer some sequences itself, such as a cursor-position
// query. Those replies go into its input pipe and are delivered to the shell by
// writeTerminalResponses, so Write can block until that goroutine reads them.
//
// The loop ends when the transport reports an error. io.EOF, or any error after
// close, is the normal end of a session. It also ends when the emulator refuses
// a write because it was closed.
func (e *terminalEngine) readProcessOutput() {
	defer e.workers.Done()
	buffer := make([]byte, 4096)
	for {
		count, err := e.transport.Read(buffer)
		// Process data before the error: a Read may return the last bytes
		// together with io.EOF.
		if count > 0 {
			e.mu.Lock()
			_, writeErr := e.emulator.Write(buffer[:count])
			e.version++
			e.mu.Unlock()
			if writeErr != nil {
				return
			}
			e.notifyChanged()
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !e.isClosed() {
				slog.Debug("terminal output reader stopped", logErrorKey, err)
			}
			return
		}
	}
}

// writeTerminalResponses is the input pump: every byte the shell receives goes
// through this loop, which is the transport's only writer. It runs on its own
// goroutine for the engine's whole life.
//
// emulator.Read returns two kinds of bytes, both of which the shell must see:
//   - key input from sendKey and sendText, already encoded by the emulator
//     for the current terminal modes (for example, arrow keys in application
//     cursor mode), and
//   - the emulator's replies to queries the shell made, such as reporting the
//     cursor position.
//
// The emulator's input is an unbuffered io.Pipe. Every SendKey or SendText
// call, and every reply produced inside emulator.Write, blocks until this loop
// reads it. If this loop is stuck writing to the transport, key input stalls,
// and so does any caller holding mu (sendKey, or readProcessOutput mid-parse).
//
// close unblocks the pending Read by sending a NUL byte after closing done.
// The isClosed check then drops it instead of writing to a closed transport.
// A failed transport write is logged and the loop keeps going; it ends when
// emulator.Read returns an error once the emulator is closed.
func (e *terminalEngine) writeTerminalResponses() {
	defer e.workers.Done()
	buffer := make([]byte, 4096)
	for {
		count, err := e.emulator.Read(buffer)
		if count > 0 {
			if e.isClosed() {
				return
			}
			if _, writeErr := e.transport.Write(buffer[:count]); writeErr != nil {
				if !e.isClosed() {
					slog.Debug("terminal input writer stopped", logErrorKey, writeErr)
				}
			}
		}
		if err != nil {
			return
		}
	}
}

// notifyChanged records that the screen changed without ever blocking the
// output pump. If a signal is already pending, Bubble Tea has not redrawn yet
// and will pick up this change too, so the extra signal is dropped.
func (e *terminalEngine) notifyChanged() {
	select {
	// if the channel can accept a notif immediately, send it;
	case e.changed <- struct{}{}:
	// otherwise, don't block (do nothing)
	default:
	}
}

// waitForChange returns a Bubble Tea command that waits for the next screen
// change and reports it as a terminalChangedMsg. The session issues a new
// command after each message, so there is always exactly one waiting. It
// returns nil once the engine closes.
func (e *terminalEngine) waitForChange() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-e.changed:
			return terminalChangedMsg{terminal: e}
		case <-e.done:
			return nil
		}
	}
}

// sendKey encodes one key press and queues it for the shell. Printable text
// is sent as is (Alt adds an ESC prefix). Control and special keys, such as
// Ctrl+C or the arrows, are encoded by the emulator for its current modes.
// It blocks until writeTerminalResponses takes the bytes (see that function),
// so callers run it from a tea.Cmd, never from Update.
func (e *terminalEngine) sendKey(message tea.KeyPressMsg) {
	key := message.Key()
	e.mu.Lock()
	defer e.mu.Unlock()
	// the bitwise AND checks if Ctrl is not pressed
	if key.Text != "" && key.Mod&tea.ModCtrl == 0 {
		// check if Alt is pressed
		if key.Mod&tea.ModAlt != 0 {
			e.emulator.SendText("\x1b")
		}
		// send normal text input to the emulator
		e.emulator.SendText(key.Text)
		return
	}
	// send special key input to the emulator
	e.emulator.SendKey(vt.KeyPressEvent(key))
}

// sendText queues raw text for the shell, with the same blocking as sendKey.
func (e *terminalEngine) sendText(input string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.emulator.SendText(input)
}

// resize tells the shell's PTY the new size first, so programs receive
// SIGWINCH and redraw, then resizes the emulator's grid to match.
func (e *terminalEngine) resize(width, height int) error {
	if err := e.transport.Resize(width, height); err != nil {
		return err
	}
	e.mu.Lock()
	e.emulator.Resize(width, height)
	e.version++
	e.mu.Unlock()
	e.notifyChanged()
	return nil
}

// view renders the screen with the active theme's text color on every cell
// that sets no foreground of its own, so default-colored text, including the
// Bash prompt, is readable on the learner's terminal background. Colors that
// programs set explicitly, and the background, are left as the shell drew them.
// The result is cached until the screen or the theme's text color changes.
func (e *terminalEngine) view() string {
	text := ui.TextColor
	r, g, b, a := text.RGBA()
	textKey := [4]uint32{r, g, b, a}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.hasCachedView && e.cachedVersion == e.version && e.cachedText == textKey {
		return e.cachedView
	}

	width, height := e.emulator.Width(), e.emulator.Height()
	if e.frame == nil || e.frame.Width() != width || e.frame.Height() != height {
		e.frame = uv.NewBuffer(width, height)
	} else {
		e.frame.Clear()
	}
	for y := range height {
		for x := 0; x < width; {
			advance := 1
			if cell := e.emulator.CellAt(x, y); cell != nil {
				themed := cell.Clone()
				if themed.Width > 1 {
					advance = themed.Width
				}
				if themed.Style.Fg == nil && themed.Content != "" && themed.Content != " " {
					themed.Style.Fg = text
				}
				e.frame.SetCell(x, y, themed)
			}
			x += advance
		}
	}

	e.cachedView = e.frame.Render()
	e.cachedVersion = e.version
	e.cachedText = textKey
	e.hasCachedView = true
	return e.cachedView
}

// cursor returns the emulator's cursor position within the terminal grid and
// whether the shell currently shows it.
func (e *terminalEngine) cursor() (int, int, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	position := e.emulator.CursorPosition()
	return position.X, position.Y, e.cursorVisible
}

func (e *terminalEngine) isClosed() bool {
	select {
	case <-e.done:
		return true
	default:
		return false
	}
}

// close stops both I/O goroutines and releases the transport and emulator.
// It is safe to call more than once; only the first call does the work.
func (e *terminalEngine) close() error {
	var closeErr error
	e.closeOnce.Do(func() {
		// Mark the engine closed before waking anything, so the goroutines see
		// done when they resume.
		close(e.done)
		// Wake writeTerminalResponses from its blocking emulator.Read. It sees
		// done and drops the byte.
		e.mu.Lock()
		e.emulator.SendText("\x00")
		e.mu.Unlock()
		// Closing the transport makes the blocked transport.Read in
		// readProcessOutput return an error.
		if err := e.transport.Close(); err != nil {
			closeErr = err
		}
		e.workers.Wait()
		// Both goroutines are gone, so nothing else uses the emulator now.
		e.mu.Lock()
		if err := e.emulator.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		e.mu.Unlock()
	})
	return closeErr
}
