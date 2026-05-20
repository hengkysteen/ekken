//go:build !windows

package passistant

import (
	"os/exec"
	"syscall"
)

func prepareCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil && cmd.Process.Pid > 0 {
		// Sending SIGKILL to -PID kills the entire process group
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
