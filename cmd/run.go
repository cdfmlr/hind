package cmd

import (
	"hind/container"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type runOptions struct {
	Tty         bool
	Interactive bool
	Image       string
	Command     []string // COMMAND ARG...
}

func runCommand() *cobra.Command {
	opts := runOptions{}

	var cmd = &cobra.Command{
		Use:   "run [flags] IMAGE [COMMAND] [ARG...]",
		Short: "Create and run a new container",
		Long:  `Create and run a new container with namespace and cgroups limit.`,
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.Image = args[0]
			if len(args) > 1 {
				opts.Command = args[1:]
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			slog.Debug("hind run",
				slog.Group("flags", "tty", opts.Tty, "interactive", opts.Interactive),
				slog.Group("args", "image", opts.Image, "command", opts.Command),
			)
			runRun(opts)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.Tty, "tty", "t", false, "Allocate a pseudo-TTY")
	flags.BoolVarP(&opts.Interactive, "interactive", "i", false, "Keep STDIN open")

	// SetInterspersed to false to support:
	//  docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
	// parse flags after IMAGE as ARGS instead of OPTIONS
	flags.SetInterspersed(false)

	return cmd
}

func runRun(opts runOptions) {
	slog.Debug("cmd: runRun", "opts", opts)
	err := container.Run(opts.Tty || opts.Interactive, opts.Command)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCommand())
}
