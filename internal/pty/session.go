package pty

import (
	"context"
	"os"
	"os/exec"
	"shellforge/internal/container"
	"sync"

	"github.com/creack/pty"
)

// Session owns one interactive Bash process and its pseudo-terminal.
type Session struct {
	file      *os.File
	command   *exec.Cmd
	container *container.LessonContainer
	closeOnce sync.Once
}

// Start launches an interactive Bash session inside a disposable Docker container.
func Start(width, height int) (*Session, error) {
	lessonContainer, err := container.CreateAndStart(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			lessonContainer.Remove(context.Background())
		}
	}()

	command := lessonContainer.ShellCommand()

	file, err := pty.StartWithSize(command, &pty.Winsize{
		Cols: terminalDimension(width, 80),
		Rows: terminalDimension(height, 24),
	})
	if err != nil {
		return nil, err
	}

	return &Session{file: file, command: command, container: lessonContainer}, nil
}

// Read reads output emitted by the shell.
func (s *Session) Read(buffer []byte) (int, error) {
	return s.file.Read(buffer)
}

// Write forwards terminal input to the shell.
func (s *Session) Write(input []byte) (int, error) {
	return s.file.Write(input)
}

// Resize updates the pseudo-terminal dimensions.
func (s *Session) Resize(width, height int) error {
	return pty.Setsize(s.file, &pty.Winsize{
		Cols: terminalDimension(width, 80),
		Rows: terminalDimension(height, 24),
	})
}

// Close terminates the shell and deletes its temporary working directory.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		_ = s.file.Close()
		if s.command.Process != nil {
			_ = s.command.Process.Kill()
		}
		_ = s.command.Wait()
		s.container.Remove(context.Background())
	})
}

func terminalDimension(value, fallback int) uint16 {
	if value <= 0 {
		return uint16(fallback)
	}
	if value > 65535 {
		return 65535
	}
	return uint16(value)
}
