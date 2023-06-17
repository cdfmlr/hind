package container

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/exp/slog"
)

// NewParentProcess creates a PID 1 process for container.
//
// It returns the command and a pipe to send init command to the process.
// The command is executed in the host.
func NewParentProcess(tty bool, cmdPipeR *os.File) (cmd *exec.Cmd) {
	cmd = exec.Command("/proc/self/exe", "init")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}

	// file descriptor 3 to receive init command
	cmd.ExtraFiles = []*os.File{cmdPipeR}

	return cmd
}

// RunContainerInitProcess bootstraps the command inside the container.
//
// It is executed as the PID 1 inside the container.
// And than core-replaced by the command.
func RunContainerInitProcess() error {
	slog.Info("[container] RunContainerInitProcess: bootstrapping container")
	command, err := recvAndCheckCommand()
	if err != nil {
		return err
	}

	setupMount()
	slog.Info("[container] pid 1 setup mount")

	return execve(command)
}

// recvAndCheckCommand wraps recvCommand.
func recvAndCheckCommand() ([]string, error) {
	command, err := recvCommand()
	if err != nil {
		slog.Error("[container] pid 1 failed to receive command", "err", err)
		return nil, err
	}

	if len(command) == 0 {
		slog.Error("[container] pid 1 received empty command")
		return nil, ErrEmptyCommand
	}
	slog.Info("[container] pid 1 received command", "command", command)

	return command, nil
}

func setupMount() {
	// 阻断 shared subtree: mount --make-rprivate /
	syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")

	// wd, err := os.Getwd()
	// if err != nil {
	// 	slog.Error("[container] RunContainerInitProcess: failed to get working directory", "err", err)
	// 	return
	// }
	// slog.Info("[container] RunContainerInitProcess: working directory", "wd", wd)
	// pivotRoot(wd)

	// 挂进程: NOEXEC: 不允许其他程序运行，NOSUID 不允许 set uid
	syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV), "")
}

// execve looks for the command and replaces the current process with it.
func execve(command []string) error {
	exe, err := exec.LookPath(command[0])
	if err != nil {
		slog.Error("[container] pid1 failed to find command", "err", err)
		return err
	}
	slog.Info("[container] pid 1 found command in path", "exe", exe)

	slog.Info("[container] pid 1 ready to execve the command. Bootstrapping done. Bye.", "command", command)
	if err := syscall.Exec(exe, command[:], os.Environ()); err != nil {
		slog.Error("[container] pid 1: execve failed", "err", err)
	}

	return nil
}
