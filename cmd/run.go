package cmd

import (
	"fmt"
	"hind/cgroups"
	"hind/container"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type runOptions struct {
	Tty         bool
	Interactive bool
	Image       string
	Command     []string // COMMAND ARG...
	Resources   cgroups.Resources
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

			// history compatibility: NOIMG -> /
			for _, noimg := range []string{"noimg"} {
				if strings.ToLower(opts.Image) == noimg {
					slog.Warn("no image specified, use / as rootfs.")
					opts.Image = "/"
				}
			}

			if _, err := os.Stat(opts.Image); err != nil {
				return fmt.Errorf("bad image: %w", err)
			}

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

	// resources
	flags.Int64Var((*int64)(&opts.Resources.CpuQuotaUs), "cpu-quota-us", 0, "The CPU hardcap limit (in usecs). Allowed cpu time in a given period.")
	flags.Uint64Var((*uint64)(&opts.Resources.CpuPeriodUs), "cpu-period-us", 0, "CPU period to be used for hardcapping (in usecs). 0 to use system default.")
	flags.StringVar((*string)(&opts.Resources.CpuSetCpus), "cpuset-cpus", "", "The requested CPUs to be used by tasks within this cgroup: 0-4,6,8-10")
	flags.Uint64Var((*uint64)(&opts.Resources.MemoryLimitBytes), "memory-limit-bytes", 0, "Memory limit in bytes")

	// SetInterspersed to false to support:
	//  docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
	// parse flags after IMAGE as ARGS instead of OPTIONS
	flags.SetInterspersed(false)

	return cmd
}

func runRun(opts runOptions) {
	slog.Info("[cmd/run] Create and run a new container.", "opts", opts)
	err := container.Run(opts.Tty || opts.Interactive, opts.Image, opts.Command, opts.Resources)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCommand())
}
