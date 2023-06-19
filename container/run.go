package container

import (
	"fmt"
	"hind/cgroups"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// Run creates a container and runs the command in it.
//
// This function is executed in the host namespace.
//
// rootDir is the root directory of the container.
func Run(container *Container) error {
	slog.Info("[host] setup and run container.", "container", container)

	if err := checkContainer(container); err != nil {
		slog.Error("[host] bad container.", "err", err)
		return err
	}

	// create pipe to send command to the container
	cmdPipeR, cmdPipeW, err := os.Pipe()
	if err != nil {
		slog.Error("[host] NewParentProcess: failed to create pipe", "err", err)
		return err
	}

	// create container process: PID 1 in the container
	containerExe := NewParentProcess(container, cmdPipeR)
	if err := containerExe.Start(); err != nil {
		slog.Error("[host] Failed to start the parent process.", "err", err)
	}
	container.Process = containerExe.Process
	slog.Info("[host] container process started.", "pid", container.Process.Pid)

	// the work dir should be cleaned up at the very end. So it is deferred first.
	defer cleanupWorkDir(container)

	// cgroup setup
	cgroupCleanup, err := setupCgroup(container)
	if err != nil {
		slog.Error("[host] Failed to setup cgroup. Kill the container.", "err", err)
		container.Process.Kill()
		return err
	}
	defer cgroupCleanup()

	// root dir setup

	container.InContainerConfig = &InContainerConfig{
		RootDir: container.WorkDir, // will be set later by setupRootDir
		Command: container.Command,
	}

	rootDirCleanup, err := setupRootDir(container)
	if err != nil {
		slog.Error("[host] Failed to setup root dir. Kill the container.", "err", err)
		container.Process.Kill()
		return err
	}
	defer rootDirCleanup()

	// send the command to the container

	sendConfig(container.InContainerConfig, cmdPipeW)
	slog.Info("[host] Command sent, closing the pipe (w).")
	cmdPipeW.Close()

	// the command is running in the container now

	state, err := container.Process.Wait()
	if err != nil {
		slog.Error("[host] container process wait failed.", "err", err)
		return err
	}

	slog.Info("[host] container process exited.", "state", state)

	return nil
}

// checkContainer errors if the container misses necessary fields.
// And sets default values for optional fields.
func checkContainer(container *Container) error {
	// necessary

	if container == nil {
		return fmt.Errorf("nil container")
	}
	if container.Command == nil || len(container.Command) < 1 {
		return ErrEmptyCommand
	}
	if container.ImagePath == "" {
		return fmt.Errorf("empty image path")
	}

	// optional

	if container.ID == "" {
		container.ID = randContainerID()
	}
	if container.Name == "" {
		container.Name = randContainerName(container.ID)
	}
	if container.WorkDir == "" {
		container.WorkDir = defaultWorkDir(container.ID)
	}
	if container.Resources == nil {
		container.Resources = &cgroups.Resources{}
	}

	return nil
}

func randContainerID() string {
	return uuid.NewString()
}

// TODO: human readable name
func randContainerName(containerID string) string {
	if len(containerID) < 8 {
		return containerID
	}
	return containerID[0:8]
}

func defaultWorkDir(containerID string) string {
	return path.Join(os.TempDir(), "hind", "container", containerID)
}

func setupCgroup(container *Container) (cleanUpFunc, error) {
	if container == nil || container.Process == nil || container.Resources == nil {
		panic("setupCgroup: invalid container.")
	}

	res := *container.Resources

	cgroupManager, err := cgroups.NewV1fsManager("/sys/fs/cgroup/", "hind/container")
	if err != nil {
		slog.Error("[host] Failed to create cgroup manager.", "err", err)
		return func() {}, err
	}
	slog.Info("[host] Cgroup manager created.", "manager", cgroupManager)

	cgroupManager.Set(res)
	cgroupManager.Apply(container.Process.Pid)

	slog.Info("[host] Cgroup setup done.", "pid", container.Process.Pid, "resources", res, "manager", cgroupManager)

	return func() {
		cgroupManager.Destroy()
		// TODO: get cgroup name
		slog.Info("[host] Cgroup destroyed.", "manager", cgroupManager)
	}, nil
}

func setupRootDir(container *Container) (cleanUpFunc, error) {
	if container == nil || container.Process == nil {
		panic("setupRootDir: invalid container.")
	}

	st, err := os.Stat(container.ImagePath)
	if err != nil {
		return func() {}, fmt.Errorf("failed to stat image path: %w", err)
	}
	if !st.IsDir() && !container.Overlay {
		slog.Warn("[host] image is not a directory. OverlayFS is forced on.", "imagePath", container.ImagePath)
		container.Overlay = true
	}

	if !container.Overlay {
		noOverlayAlert(container)
		container.InContainerConfig.RootDir = container.ImagePath
		return func() {}, nil
	}

	// setup overlayfs

	if err := makeOverlayFS(container); err != nil {
		return func() {}, fmt.Errorf("failed to make overlayfs: %w", err)
	}
	container.InContainerConfig.RootDir = container.overlayMergedDir()
	slog.Info("[host] OverlayFS setup done. InContainerConfig.RootDir -> overlayMergedDir", "mergedDir", container.InContainerConfig.RootDir)

	return func() {
		destroyOverlayFS(container)
		slog.Info("[host] cleanupOverlayFS: overlayfs destroyed.", "removed", container.overlayRootDir())
	}, nil
}

func cleanupWorkDir(container *Container) {
	if container != nil && container.WorkDir == defaultWorkDir(container.ID) {
		if err := os.RemoveAll(container.WorkDir); err != nil {
			slog.Error("[host] Failed to remove the tmp work dir.", "err", err)
		}
	}

	slog.Info("[host] cleanupWorkDir: tmp work dir cleanup.", "removed", container.WorkDir)
}

// cleanUpFunc is context.CancelFunc
type cleanUpFunc func()

// TODO: mv this to cmd package. This is not a host function, but a ux design.
func noOverlayAlert(container *Container) {
	pauseTime := 1 * time.Second
	slog.Warn("[host] No image found. "+
		"The container will be run in the image directory directly "+
		"and it could modify anything in this directory. "+
		"This could be dangerous."+
		"We seriously recommend you to think twice before you continue. "+
		fmt.Sprintf("The program is now pausing for %v... Press Ctrl+C to abort.", pauseTime),
		"ImagePath", container.ImagePath)
	time.Sleep(pauseTime)
}
