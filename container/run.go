package container

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/exp/slog"
)

func NewParentProcess(tty bool, command []string) *exec.Cmd {
	args := append([]string{"init"}, command...)
	cmd := exec.Command("/proc/self/exe", args...)

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

	return cmd
}

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

func RunContainerInitProcess(command []string) error {
	if len(command) < 1 {
		slog.Error("[container] RunContainerInitProcess: empty command, nothing to do.")
		return ErrEmptyCommand
	}

	slog.Info("[container] RunContainerInitProcess", "command", command)

	// 阻断 shared subtree: mount --make-rprivate /
	syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")

	// 挂进程: NOEXEC: 不允许其他程序运行，NOSUID 不允许 set uid
	syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV), "")

	if err := syscall.Exec(command[0], command[:], os.Environ()); err != nil {
		slog.Error("[container] RunContainerInitProcess: execve failed", "err", err)
	}

	return nil
}

var (
	ErrEmptyCommand = errors.New("empty command")
)
