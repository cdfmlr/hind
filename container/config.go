package container

import (
	"encoding/json"
	"fmt"
	"hind/cgroups"
	"io"
	"os"

	"golang.org/x/exp/slog"
)

// Container is the host's view of a container.
//
// maybe i should make it a ContainerContext?
type Container struct {
	// Metadata

	ID   string
	Name string

	// InContainerConfig's blueprint

	Command []string

	// Setup config

	WorkDir   string // WorkDir is a dir to do the setup work. NOT the $(pwd) of the container.
	TTY       bool
	ImagePath string // directory | tar file
	Overlay   bool   // if true, use overlayfs to make the image read-only
	Resources *cgroups.Resources

	// Runtime config

	Process           *os.Process        // the process of the container
	InContainerConfig *InContainerConfig // the config sent to the container
	OverlayConfig     *overlayConfig     // the config of the overlayfs
}

// InContainerConfig is the configuration to initialize a container.
//
// It is sent from the host to the container through a pipe.
// PID 1 in the container sets up the container according to it before exec the command.
//
// This is exported for json encoding.
type InContainerConfig struct {
	RootDir string
	Command []string
}

// sendConfig writes the InContainerConfig to the pipe.
// The command is json encoded.
//
// This function is executed in the host.
// Use recvConfig() in the container to read the sent command.
func sendConfig(config *InContainerConfig, cmdPipeW io.Writer) error {
	if config == nil {
		return fmt.Errorf("sendCommand: config is nil")
	}
	slog.Debug("sendCommand: Sending command", "config", config)
	return json.NewEncoder(cmdPipeW).Encode(config)
}

// recvConfig reads the InContainerConfig sent by sendConfig() through the pipe.
// The received config should be json encoded.
//
// This function is executed in the container.
func recvConfig() (*InContainerConfig, error) {
	pipe := os.NewFile(uintptr(3), "pipe")

	cfgJson, err := io.ReadAll(pipe)
	if err != nil {
		slog.Debug("recvCommand: read command from pipe error", "err", err)
		return nil, err
	}
	slog.Debug("recvCommand: Command received", "cmd", string(cfgJson))

	var config InContainerConfig
	json.Unmarshal(cfgJson, &config)

	return &config, nil
}
