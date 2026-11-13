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

type terminalTransport interface {
	io.ReadWriteCloser
	Resize(width, height int) error
}

type terminalChangedMsg struct {
	terminal *terminalEngine
}

type terminalEngine struct {
	mu            sync.RWMutex
	emulator      *vt.Emulator
	transport     terminalTransport
	changed       chan struct{}
	done          chan struct{}
	cursorVisible bool
	closeOnce     sync.Once
	workers       sync.WaitGroup

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

func (e *terminalEngine) readProcessOutput() {
	defer e.workers.Done()
	buffer := make([]byte, 4096)
	for {
		count, err := e.transport.Read(buffer)
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

func (e *terminalEngine) notifyChanged() {
	select {
	// if the channel can accept a notif immediately, send it;
	case e.changed <- struct{}{}:
	// otherwise, don't block (do nothing)
	default:
	}
}

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

func (e *terminalEngine) sendText(input string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.emulator.SendText(input)
}

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

func (e *terminalEngine) close() error {
	var closeErr error
	e.closeOnce.Do(func() {
		close(e.done)
		e.mu.Lock()
		e.emulator.SendText("\x00")
		e.mu.Unlock()
		if err := e.transport.Close(); err != nil {
			closeErr = err
		}
		e.workers.Wait()
		e.mu.Lock()
		if err := e.emulator.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
		e.mu.Unlock()
	})
	return closeErr
}
