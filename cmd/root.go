package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/prompt"
	"github.com/decampsrenan/spm/internal/resolver"
	"github.com/decampsrenan/spm/internal/runner"
)

var dryRun bool

var rootCmd = &cobra.Command{
	Use:   "spm",
	Short: "Smart Package Manager — auto-detects npm/yarn/pnpm and proxies commands",
	// Running `spm` with no args is equivalent to `spm install`
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("install", args)
	},
	// Silence Cobra's default error/usage printing so we control output
	SilenceUsage:  true,
	SilenceErrors: true,
}

var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("install", args)
	},
}

var addCmd = &cobra.Command{
	Use:   "add [packages...]",
	Short: "Add one or more packages",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("add", args)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print command instead of executing it")
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(addCmd)
}

func Execute() {
	// If the command is not recognized by Cobra, treat it as a script run (fallback)
	// We do this by intercepting unknown subcommands via a custom args function.
	rootCmd.Args = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// Override the default behavior: if Cobra can't find the subcommand,
	// run it as a script.
	if len(os.Args) > 1 {
		first := os.Args[1]
		if first != "install" && first != "i" && first != "add" &&
			first != "help" && first != "completion" &&
			first != "--help" && first != "-h" &&
			first != "--dry-run" {
			// Treat as a script fallback
			rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
				// args from cobra includes "first" as the first element; skip it
				var extra []string
				if len(args) > 0 {
					extra = args[1:]
				}
				return run(first, extra)
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(command string, extraArgs []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}

	detections, err := detector.Detect(cwd)
	if err != nil {
		return err
	}

	var det detector.Detection
	if len(detections) == 1 {
		det = detections[0]
	} else {
		det, err = prompt.Select(detections)
		if err != nil {
			return err
		}
	}

	args := resolver.Resolve(det.PM, command, extraArgs)
	return runner.Run(args, dryRun)
}
