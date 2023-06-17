package container

import (
	"encoding/json"
	"io"
	"os"

	"golang.org/x/exp/slog"
)

// sendCommand writes the init command to the pipe.
// The command is json encoded:
//
//	["/bin/sh", "-c", "echo hello"]
//
// This function is executed in the host.
// Use recvCommand() in the container to read the sent command.
func sendCommand(command []string, cmdPipeW io.Writer) error {
	return json.NewEncoder(cmdPipeW).Encode(command)
}

// recvCommand reads the command sent by sendCommand() through the pipe.
// The received command should be json encoded:
//
//	["/bin/sh", "-c", "echo hello"]
//
// This function is executed in the container.
func recvCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3), "pipe")

	cmdJson, err := io.ReadAll(pipe)
	if err != nil {
		slog.Error("read command from pipe error", "err", err)
		return nil, err
	}
	slog.Debug("[container] Command received", "cmd", string(cmdJson))

	var cmd []string
	json.Unmarshal(cmdJson, &cmd)

	return cmd, nil
}
