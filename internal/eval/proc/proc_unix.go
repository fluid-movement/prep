//go:build !windows

// Package proc starts eval commands in their own process group and kills the
// whole group, so a harness's grandchildren end with it.
package proc

import (
	"os/exec"
	"syscall"
)

// Group makes cmd start in a new process group.
func Group(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Kill kills the started cmd's process group.
func Kill(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
