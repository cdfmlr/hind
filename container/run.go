package container

import (
	"errors"
	"hind/cgroups"
	"os"

	"golang.org/x/exp/slog"
)

// Run creates a container and runs the command in it.
//
// This function is executed in the host namespace.
func Run(tty bool, command []string, res cgroups.Resources) error {
	slog.Info("[host] Run command in container.", "tty", tty, "command", command, "resources", res)
	if len(command) < 1 {
		slog.Error("[host] Run: empty command, nothing to do.")
		return ErrEmptyCommand
	}

	// create pipe to send command to the container
	cmdPipeR, cmdPipeW, err := os.Pipe()
	if err != nil {
		slog.Error("[host] NewParentProcess: failed to create pipe", "err", err)
		return err
	}

	// create container process: PID 1 in the container
	container := NewParentProcess(tty, cmdPipeR)
	if err := container.Start(); err != nil {
		slog.Error("[host] Failed to start the parent process.", "err", err)
	}
	slog.Info("[host] container process started.", "pid", container.Process.Pid)

	// cgroup setup
	// I am discouraged to put this here. But I cannot wrap it with a function
	// due to the defer & killing job.
	cgroupManager, err := cgroups.NewV1fsManager("/sys/fs/cgroup/", "hind/container")
	if err != nil {
		slog.Error("[host] Failed to create cgroup manager. kill the process.", "err", err)
		container.Process.Kill()
		return err
	}
	defer cgroupManager.Destroy()
	slog.Info("[host] Cgroup manager created.", "manager", cgroupManager)

	cgroupManager.Set(res)
	cgroupManager.Apply(container.Process.Pid)

	slog.Info("[host] Cgroup setup done.", "pid", container.Process.Pid, "resources", res, "manager", cgroupManager)

	// send the command
	sendCommand(command, cmdPipeW)
	slog.Info("[host] Command sent, closing the pipe (w).")
	cmdPipeW.Close()

	// the command is running in the container now
	container.Wait()
	slog.Info("[host] container process exited.")

	return nil
}

var (
	ErrEmptyCommand = errors.New("empty command")
)
