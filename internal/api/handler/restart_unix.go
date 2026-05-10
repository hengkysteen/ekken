//go:build !windows

package handler

import (
	"os"
	"syscall"
)

func (h *Handler) restartApp() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	args := os.Args
	env := os.Environ()

	return syscall.Exec(self, args, env)
}
