//go:build windows

// Package proc starts eval commands in their own process group and kills the
// whole group, so a harness's grandchildren end with it.
package proc

import "os/exec"

// Group does nothing on Windows: evals run on Unix only, and the binary
// only has to build.
func Group(cmd *exec.Cmd) {}

// Kill kills the started cmd's process; its children may outlive it.
func Kill(cmd *exec.Cmd) {
	cmd.Process.Kill()
}
