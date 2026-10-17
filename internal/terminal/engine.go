package terminal

import (
	"errors"
	"io"
	"log/slog"
	"sync"

	tea "charm.land/bubbletea/v2"
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
}

func newTerminalEngine(width, height int, transport terminalTransport) *terminalEngine {
	engine := &terminalEngine{
		emulator:      vt.NewEmulator(width, height),
		transport:     transport,
		changed:       make(chan struct{}, 1),
		done:          make(chan struct{}),
		cursorVisible: true,
	}
	engine.emulator.SetCallbacks(vt.Callbacks{
		CursorVisibility: func(visible bool) {
			engine.cursorVisible = visible
		},
	})
	engine.workers.Add(2)
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
	case e.changed <- struct{}{}:
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
	if key.Text != "" && key.Mod&tea.ModCtrl == 0 {
		if key.Mod&tea.ModAlt != 0 {
			e.emulator.SendText("\x1b")
		}
		e.emulator.SendText(key.Text)
		return
	}
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
	e.mu.Unlock()
	e.notifyChanged()
	return nil
}

func (e *terminalEngine) view() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.emulator.Render()
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
