package cmd

import (
	"hind/container"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func initCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: "run exec a command (read from 3) in container (interal use only! do not call it)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runBoot()
		},
	}

	flags := cmd.Flags()

	// SetInterspersed to false to support:
	//  docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
	// parse flags after IMAGE as ARGS instead of OPTIONS
	flags.SetInterspersed(false)

	return cmd
}

func runBoot() {
	slog.Info("[cmd/init] Booting container...")
	err := container.RunContainerInitProcess()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCommand())
}
