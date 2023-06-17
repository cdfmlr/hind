package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"golang.org/x/exp/slog"
)

// overlayFSImageLayerPath returns the path to put the extracted image layer (read-only).
func (c ConatinerConfig) overlayFSImageLayerPath() string {
	return path.Join(c.RootDir, "hind-image-"+c.ContainerID)
}

// overlayFSWritableLayerPath returns the path to put the writable layer.
func (c ConatinerConfig) overlayFSWritableLayerPath() string {
	return path.Join(c.RootDir, "hind-writable-"+c.ContainerID)
}

func (c ConatinerConfig) overlayFSMountPath() string {
	return path.Join(c.RootDir, "hind-mount-"+c.ContainerID)
}

// overlayFSMountWorkDir is used to prepare files as they are switched between the layers.
func (c ConatinerConfig) overlayFSMountWorkDir() string {
	return path.Join(c.RootDir, "hind-work-"+c.ContainerID)
}

// initOverlayFS create the overlay filesystem for the container.
// That contains:
//   - lower directory: read-only image layer
//   - upper directory: read-write container layer
//
// Overlayfs has been in the Linux kernel since 3.18.
//
// References:
//   - https://wiki.archlinux.org/title/Overlay_filesystem (arch wiki yyds)
func initOverlayFS(config *ConatinerConfig) error {
	err := extractImage(config.ImagePath, config.overlayFSImageLayerPath())
	if err != nil {
		slog.Error("[container] initOverlayFS: error extracting image", "err", err)
		return fmt.Errorf("error extracting image: %w", err)
	}

	err = os.MkdirAll(config.overlayFSWritableLayerPath(), 0755)
	if err != nil {
		slog.Error("[container] initOverlayFS: error creating writable layer", "err", err)
		return fmt.Errorf("error creating writable layer: %w", err)
	}

	err = os.MkdirAll(config.overlayFSMountWorkDir(), 0755)
	if err != nil {
		slog.Error("[container] initOverlayFS: error creating overlay work dir", "err", err)
		return fmt.Errorf("error creating overlay work dir: %w", err)
	}

	err = os.MkdirAll(config.overlayFSMountPath(), 0755)
	if err != nil {
		slog.Error("[container] initOverlayFS: error creating mount point", "err", err)
		return fmt.Errorf("error creating mount point: %w", err)
	}

	// mount overlayfs
	err = mountOverlayFS(config.overlayFSImageLayerPath(), config.overlayFSWritableLayerPath(), config.overlayFSMountWorkDir(), config.overlayFSMountPath())
	if err != nil {
		slog.Error("[container] initOverlayFS: error mounting overlayfs", "err", err)
		return fmt.Errorf("error mounting overlayfs: %w", err)
	}

	return nil
}

// extractImage extracts the image layer to the target path.
func extractImage(imagePath string, targetPath string) error {
	slog.Info("[container] extracting image", "imagePath", imagePath, "targetPath", targetPath)

	// image file should exist
	if _, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("image file %s does not exist", imagePath)
	}

	// targetPath should be empty
	if l, err := os.ReadDir(targetPath); len(l) != 0 {
		return fmt.Errorf("extractImage: target path %s is not empty", targetPath)
	} else if os.IsNotExist(err) {
		slog.Info("[container] extractImage: target path does not exist, creating...", "targetPath", targetPath)
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("error creating mount point: %w", err)
		}
	}

	// tar -xvf $imagePath -C $mountPath
	tarCmd := exec.Command("tar", "-xvf", imagePath, "-C", targetPath)
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("error extracting image: %w", err)
	}

	return nil
}

// mountOverlayFS mounts the overlay filesystem:
//
//	mount -t overlay overlay -o lowerdir=$lowerdir,upperdir=$upperdir,workdir=$workdir $mountpoint
func mountOverlayFS(lowerdir string, upperdir string, workdir string, mountpoint string) error {
	mountCmd := exec.Command(
		"mount", "-t", "overlay", "overlay",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upperdir, workdir),
		mountpoint)

	if err := mountCmd.Run(); err != nil {
		return fmt.Errorf("error mounting overlayfs: %w", err)
	}

	return nil
}

// unmountOverlayFS unmounts the overlay filesystem.
//
//	umount $mountpoint
//
// Do not call this function directly, use cleanupOverlayFS instead.
func unmountOverlayFS(mountpoint string) error {
	umountCmd := exec.Command("umount", mountpoint)
	if err := umountCmd.Run(); err != nil {
		return fmt.Errorf("error unmounting overlayfs: %w", err)
	}

	return nil
}

// cleanupOverlayFS cleans up the overlay filesystem:
//   - unmount overlayfs
//   - remove mount point, tmp work dir and the writable layer
func cleanupOverlayFS(config *ConatinerConfig) error {
	err := unmountOverlayFS(config.overlayFSMountPath())
	if err != nil {
		slog.Error("[container] cleanupOverlayFS: error unmounting overlayfs", "err", err)
		return fmt.Errorf("error unmounting overlayfs: %w", err)
	}

	err = os.RemoveAll(config.overlayFSMountPath())
	if err != nil {
		slog.Error("[container] cleanupOverlayFS: error removing mount point", "err", err)
		return fmt.Errorf("error removing mount point: %w", err)
	}

	err = os.RemoveAll(config.overlayFSMountWorkDir())
	if err != nil {
		slog.Error("[container] cleanupOverlayFS: error removing overlay work dir", "err", err)
		return fmt.Errorf("error removing overlay work dir: %w", err)
	}

	err = os.RemoveAll(config.overlayFSWritableLayerPath())
	if err != nil {
		slog.Error("[container] cleanupOverlayFS: error removing writable layer", "err", err)
		return fmt.Errorf("error removing writable layer: %w", err)
	}

	err = os.RemoveAll(config.overlayFSImageLayerPath())
	if err != nil {
		slog.Error("[container] cleanupOverlayFS: error removing image layer", "err", err)
		return fmt.Errorf("error removing image layer: %w", err)
	}

	return nil
}

// YES, these fucking mess is generated by AI. Sue me. >_<
