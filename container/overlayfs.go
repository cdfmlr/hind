package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog"
)

// -- helper methods: paths for overlayfs --

type overlayConfig = Container

// overlayRootDir is a tmp dir to create & mount the overlay filesystem.
func (c overlayConfig) overlayRootDir() string {
	return path.Join(c.WorkDir, "/overlay-"+c.ID)
}

// overlayLowerDir is the lower directory (read-only image layer)
//
// if the image path is a directory, use it as the lower dir.
// otherwise, the image path is a tar file, extract it to the tmp overlay root dir.
func (c overlayConfig) overlayLowerDir() string {
	// if the image path is a directory, just use it as the lower dir:
	// avoid copying the dir as well as deleting it after the container exits
	if st, err := os.Stat(c.ImagePath); err == nil && st.IsDir() {
		// special case: the image path is a subdirectory of the tmp overlay root dir???
		// i have no idea who would do this and I don't know how to handle it. just warn the user.
		absImage, _ := filepath.Abs(c.ImagePath)
		absRoot, _ := filepath.Abs(c.overlayRootDir())
		if strings.HasPrefix(absImage, absRoot) {
			panic(fmt.Sprintf("image path %q is a subdirectory of the overlay root dir %q", absImage, absRoot))
		}

		return c.ImagePath
	}
	return path.Join(c.overlayRootDir(), "/image")
}

// overlayUpperDir is the upper directory (read-write container layer)
func (c overlayConfig) overlayUpperDir() string {
	return path.Join(c.overlayRootDir(), "/write")
}

// overlayMergedDir is the mount point of the overlay filesystem.
func (c overlayConfig) overlayMergedDir() string {
	return path.Join(c.overlayRootDir(), "/merge")
}

// overlayWorkDir is used to prepare files as they are switched between the layers.
func (c overlayConfig) overlayWorkDir() string {
	return path.Join(c.overlayRootDir(), "/.work")
	// 其实他会自己在 workdir=/work 里面新建一个 /work/work 来作为工作目录，所以，其实后面 .work 可以不用加
}

// -- make an overlayfs --

// makeOverlayFS create the overlay filesystem for the container.
// That contains:
//   - lower directory: read-only image layer
//   - upper directory: read-write container layer
//
// ATTENTION:
//   - this function works with the destroyOverlayFS function.
//     The following "we" refers to both of them;
//   - we read config.ID, config.ImagePath and config.WorkDir;
//   - the config.ImagePath is either a directory or a tar file:
//     (a) if it is a directory, we use it as the lower dir.
//     We do not copy, modify or delete it. And the Linux overlay
//     file system guarantees that this directory is not modified
//     by the container.
//     (b) if it is a tar file, makeOverlayFS extract it to the
//     {config.WorkDir}/overlay-{config.ID}/image directory,
//     and use this directory as the lower dir.
//     It's destroyOverlayFS's responsibility to delete this directory
//     after the container exits.
//
// References:
//   - https://wiki.archlinux.org/title/Overlay_filesystem (arch wiki yyds)
func makeOverlayFS(config *overlayConfig) error {
	if config.ImagePath != config.overlayLowerDir() {
		err := extractImage(config.ImagePath, config.overlayLowerDir())
		if err != nil {
			slog.Error("[host] initOverlayFS: error extracting image", "err", err)
			return fmt.Errorf("error extracting image: %w", err)
		}
	}

	err := os.MkdirAll(config.overlayUpperDir(), 0755)
	if err != nil {
		slog.Error("[host] initOverlayFS: error creating writable layer", "err", err)
		return fmt.Errorf("error creating writable layer: %w", err)
	}

	err = os.MkdirAll(config.overlayWorkDir(), 0755)
	if err != nil {
		slog.Error("[host] initOverlayFS: error creating overlay work dir", "err", err)
		return fmt.Errorf("error creating overlay work dir: %w", err)
	}

	err = os.MkdirAll(config.overlayMergedDir(), 0755)
	if err != nil {
		slog.Error("[host] initOverlayFS: error creating mount point", "err", err)
		return fmt.Errorf("error creating mount point: %w", err)
	}

	// mount overlayfs
	err = mountOverlayFS(config.overlayLowerDir(), config.overlayUpperDir(), config.overlayWorkDir(), config.overlayMergedDir())
	if err != nil {
		slog.Error("[host] initOverlayFS: error mounting overlayfs", "err", err)
		return fmt.Errorf("error mounting overlayfs: %w", err)
	}

	return nil
}

// extractImage extracts the image layer to the target path.
func extractImage(imagePath string, targetPath string) error {
	slog.Info("[host] extracting image", "imagePath", imagePath, "targetPath", targetPath)

	// image file should exist and not be a directory
	if statImg, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("image file %s does not exist", imagePath)
	} else if statImg.IsDir() {
		return fmt.Errorf("image file %s is a directory", imagePath)
	}

	// targetPath should be empty
	if l, err := os.ReadDir(targetPath); len(l) != 0 {
		return fmt.Errorf("extractImage: target path %s is not empty", targetPath)
	} else if os.IsNotExist(err) {
		slog.Info("[host] extractImage: target path does not exist, creating...", "targetPath", targetPath)
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

	slog.Info("DBG [host] mounting overlayfs.", "command", mountCmd.String())

	if err := mountCmd.Run(); err != nil {
		return fmt.Errorf("error mounting overlayfs: %w", err)
	}

	return nil
}

// -- destroy an overlayfs --

// destroyOverlayFS cleans up the overlay filesystem:
//   - unmount overlayfs
//   - remove mount point, tmp work dir and the writable layer
func destroyOverlayFS(config *overlayConfig) error {
	err := unmountOverlayFS(config.overlayMergedDir())
	if err != nil {
		slog.Error("[host] cleanupOverlayFS: error unmounting overlayfs", "err", err)
		return fmt.Errorf("error unmounting overlayfs: %w", err)
	}

	err = os.RemoveAll(config.overlayRootDir())
	if err != nil {
		slog.Error("[host] cleanupOverlayFS: error removing overlay tmp dir", "err", err)
		return fmt.Errorf("error removing mount point: %w", err)
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
