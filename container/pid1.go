package container

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/google/uuid"
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
	slog.Info("[container] pid 1: bootstrapping...")
	config, err := recvAndCheckConfig()
	if err != nil {
		return err
	}

	setupMount(config.RootDir)
	slog.Info("[container] pid 1 setup mount.")

	return execve(config.Command)
}

// recvAndCheckConfig wraps recvCommand.
func recvAndCheckConfig() (*ConatinerConfig, error) {
	config, err := recvConfig()
	if err != nil {
		slog.Error("[container] pid 1 failed to receive config.", "err", err)
		return nil, err
	}
	if config == nil {
		slog.Error("[container] pid 1 received nil config. THIS SHOULD NOT HAPPEN.")
		return nil, ErrNilConfig
	}

	if config.RootDir == "" {
		slog.Error("[container] pid 1 received empty root dir.")
		return nil, ErrEmptyRootDir
	}

	if len(config.Command) == 0 {
		slog.Error("[container] pid 1 received empty config.")
		return nil, ErrEmptyCommand
	}
	slog.Info("[container] pid 1 received config.", "config", *config)

	return config, nil
}

func setupMount(rootDir string) {
	// 阻断 shared subtree: mount --make-rprivate /
	syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")

	slog.Info("[container] pid 1 pivot root.", "rootDir", rootDir)
	pivotRoot(rootDir)

	// I am not sure if this is necessary after a pivot_root
	syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")

	// 挂进程: NOEXEC: 不允许其他程序运行，NOSUID 不允许 set uid
	syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

// pivotRoot changes the root file system to the path newRoot.
// And make old root (the / of host) inaccessible.
func pivotRoot(newRoot string) error {
	// 0. Original:
	//	host root: /
	//	newRoot  : /path/to/image/root/

	// remounting newroot again using bind mount:
	// ensure that the current root’s old root and new root are not in the same file system
	if err := syscall.Mount(newRoot, newRoot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount rootfs error: %v", err)
	}

	// 1. Make a directory (in the newRoot) to put the old root:
	//	hostRoot : /path/to/image/root/.hostroot

	randNameForHostRoot := ".hostroot-" + uuid.NewString()
	// putOldRootHere is a path to put the old (host) root
	putOldRootHere := path.Join(newRoot, randNameForHostRoot)
	if err := os.Mkdir(putOldRootHere, 0777); err != nil {
		return err
	}

	// 2. system call pivot_root(newRoot, hostRoot):
	//
	//	host root (old /) -> /path/to/image/root/.hostroot
	//	container root (/path/to/image/root/) -> new /

	// "/": host root -> container root
	// and put the old root (host root) at the hostRoot
	if err := syscall.PivotRoot(newRoot, putOldRootHere); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	// 3. After pivot_root, the old root is mounted at /path/to/image/root/.hostroot
	// and we are in the new root (/) (original /path/to/image/root/)

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / error %v", err)
	}

	// 4. Finally, unmount the old root (mounted at /path/to/image/root/.hostroot).

	oldRootInNewRoot := path.Join("/", randNameForHostRoot)

	if err := syscall.Unmount(oldRootInNewRoot, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	return os.Remove(oldRootInNewRoot)
}

// execve looks for the command and replaces the current process with it.
func execve(command []string) error {
	exe, err := exec.LookPath(command[0])
	if err != nil {
		slog.Error("[container] pid1 failed to find command.", "err", err)
		return err
	}
	slog.Info("[container] pid 1 found command in path.", "exe", exe)

	slog.Info("[container] pid 1 ready to execve the command. Bootstrapping done. Bye.", "command", command)
	if err := syscall.Exec(exe, command[:], os.Environ()); err != nil {
		slog.Error("[container] pid 1: execve failed.", "err", err)
	}

	return nil
}

var (
	ErrEmptyCommand = errors.New("empty command")
	ErrEmptyRootDir = errors.New("empty root dir")
	ErrNilConfig    = errors.New("nil config")
)
