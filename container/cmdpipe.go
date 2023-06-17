package container

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/exp/slog"
)

// ConatinerConfig is the configuration to initialize a container.
// It is sent from the host to the container through a pipe.
type ConatinerConfig struct {
	RootDir string
	Command []string
}

// sendConfig writes the init command to the pipe.
// The command is json encoded:
//
//	["/bin/sh", "-c", "echo hello"]
//
// This function is executed in the host.
// Use recvCommand() in the container to read the sent command.
func sendConfig(config *ConatinerConfig, cmdPipeW io.Writer) error {
	if config == nil {
		return fmt.Errorf("sendCommand: config is nil.")
	}
	slog.Debug("sendCommand: Sending command", "config", config)
	return json.NewEncoder(cmdPipeW).Encode(config)
}

// recvConfig reads the command sent by sendCommand() through the pipe.
// The received command should be json encoded:
//
//	["/bin/sh", "-c", "echo hello"]
//
// This function is executed in the container.
func recvConfig() (*ConatinerConfig, error) {
	pipe := os.NewFile(uintptr(3), "pipe")

	cfgJson, err := io.ReadAll(pipe)
	if err != nil {
		slog.Debug("recvCommand: read command from pipe error", "err", err)
		return nil, err
	}
	slog.Debug("recvCommand: Command received", "cmd", string(cfgJson))

	var config ConatinerConfig
	json.Unmarshal(cfgJson, &config)

	return &config, nil
}
