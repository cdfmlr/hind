package main

import (
	"hind/cmd"
	"os"

	"golang.org/x/exp/slog"
)

const Debug = false

func init() {
	if Debug {
		var programLevel = new(slog.LevelVar)

		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
		slog.SetDefault(slog.New(h))

		programLevel.Set(slog.LevelDebug)
	}
}

func main() {
	cmd.Execute()
}
