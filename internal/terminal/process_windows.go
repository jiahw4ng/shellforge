//go:build windows

package terminal

import "os/exec"

func preparePTYCommand(_ *exec.Cmd) {}
