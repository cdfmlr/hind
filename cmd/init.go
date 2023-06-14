package cmd

import (
	"hind/container"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type initOptions struct {
	Command []string // COMMAND ARG...
}

func initCommand() *cobra.Command {
	opts := initOptions{}

	var cmd = &cobra.Command{
		Use:   "init [flags] [COMMAND] [ARG...]",
		Short: "run a command in container (interal use only! do not call it)",
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.Command = args[:]
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runBoot(opts)
		},
	}

	flags := cmd.Flags()

	// SetInterspersed to false to support:
	//  docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
	// parse flags after IMAGE as ARGS instead of OPTIONS
	flags.SetInterspersed(false)

	return cmd
}

func runBoot(opts initOptions) {
	slog.Debug("cmd: runBoot", "command", opts.Command)
	err := container.RunContainerInitProcess(opts.Command)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCommand())
}
