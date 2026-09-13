package pty

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/creack/pty"
)

// Session owns one interactive Bash process and its pseudo-terminal.
type Session struct {
	file      *os.File
	command   *exec.Cmd
	directory string
	closeOnce sync.Once
}

// Start launches a clean interactive Bash session in a temporary directory.
func Start(width, height int) (*Session, error) {
	directory, err := os.MkdirTemp("", "shellforge-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(directory)
		}
	}()

	binDirectory, err := createDisabledNano(directory)
	if err != nil {
		return nil, err
	}

	command := exec.Command("/bin/bash", "--noprofile", "--norc", "-i")
	command.Dir = directory
	command.Env = shellEnvironment(directory, binDirectory)

	file, err := pty.StartWithSize(command, &pty.Winsize{
		Cols: terminalDimension(width, 80),
		Rows: terminalDimension(height, 24),
	})
	if err != nil {
		return nil, err
	}

	return &Session{file: file, command: command, directory: directory}, nil
}

func createDisabledNano(directory string) (string, error) {
	binDirectory := filepath.Join(directory, "bin")
	if err := os.Mkdir(binDirectory, 0o755); err != nil {
		return "", err
	}

	stub := "#!/bin/sh\nprintf '%s\\n' 'nano is disabled in Shellforge.' >&2\nexit 127\n"
	if err := os.WriteFile(filepath.Join(binDirectory, "nano"), []byte(stub), 0o755); err != nil {
		return "", err
	}

	return binDirectory, nil
}

func shellEnvironment(directory, binDirectory string) []string {
	filtered := make([]string, 0, len(os.Environ())+3)
	for _, setting := range os.Environ() {
		if strings.HasPrefix(setting, "HOME=") || strings.HasPrefix(setting, "PATH=") ||
			strings.HasPrefix(setting, "PS1=") || strings.HasPrefix(setting, "TERM=") {
			continue
		}
		filtered = append(filtered, setting)
	}

	return append(filtered,
		"HOME="+directory,
		"PATH="+binDirectory+":"+os.Getenv("PATH"),
		"PS1=shellforge$ ",
		"TERM=xterm-256color",
	)
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
		_ = os.RemoveAll(s.directory)
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
