//go:build !unix

package command

import (
	"os"
	"os/exec"
)

func configureProcessGroup(*exec.Cmd) {}

func terminateProcessGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	return process.Kill()
}
