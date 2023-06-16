package container

import (
	"errors"

	"golang.org/x/exp/slog"
)

// Run creates a container and runs the command in it.
//
// This function is executed in the host namespace.
func Run(tty bool, command []string) error {
	if len(command) < 1 {
		slog.Error("[container] Run: empty command, nothing to do.")
		return ErrEmptyCommand
	}

	slog.Info("[container] Run", "tty", tty, "command", command)

	parent := NewParentProcess(tty, command)

	slog.Info("[container] NewParentProcess: exec", "command", parent.Args)
	if err := parent.Start(); err != nil {
		slog.Error("[container] Failed to start the parent process", "err", err)
	}
	parent.Wait()
	return nil
}

var (
	ErrEmptyCommand = errors.New("empty command")
)
