//go:build windows

package handler

import (
	"os"
	"os/exec"
)

func (h *Handler) restartApp() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	args := os.Args
	env := os.Environ()

	cmd := exec.Command(self, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env

	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
