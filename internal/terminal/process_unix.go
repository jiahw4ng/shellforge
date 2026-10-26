//go:build !windows

package terminal

import (
	"os/exec"
	"syscall"
)

func preparePTYCommand(command *exec.Cmd) {
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.Setctty = true
	command.SysProcAttr.Setsid = true
}
