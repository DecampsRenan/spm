package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/prompt"
	"github.com/decampsrenan/spm/internal/resolver"
	"github.com/decampsrenan/spm/internal/runner"
)

var dryRun bool
var vibes bool
var notify bool

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

var removeCmd = &cobra.Command{
	Use:   "remove [packages...]",
	Short: "Remove one or more packages",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("remove", args)
	},
}

var playSoundCmd = &cobra.Command{
	Use:    "_play-sound [name]",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return audio.PlaySound(args[0])
	},
}

var playMusicCmd = &cobra.Command{
	Use:    "_play-music [fade-in-seconds]",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secs, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid fade-in duration: %w", err)
		}
		return audio.PlayMusicAndWait(time.Duration(secs) * time.Second)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print command instead of executing it")
	rootCmd.PersistentFlags().BoolVar(&vibes, "vibes", false, "Play background music during install")
	rootCmd.PersistentFlags().BoolVar(&notify, "notify", false, "Play a sound when the command finishes")
	// Allow unknown flags to pass through to the underlying package manager
	// (e.g. spm add react --save-dev, spm dev --port 3000)
	rootCmd.FParseErrWhitelist.UnknownFlags = true
	addCmd.FParseErrWhitelist.UnknownFlags = true
	removeCmd.FParseErrWhitelist.UnknownFlags = true
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(playSoundCmd)
	rootCmd.AddCommand(playMusicCmd)
}

func SetVersion(v string) {
	rootCmd.Version = v
	rootCmd.Flags().BoolP("version", "v", false, "Print the version")
}

func Execute() {
	// If the command is not recognized by Cobra, treat it as a script run (fallback)
	// We do this by intercepting unknown subcommands via a custom args function.
	rootCmd.Args = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// Override the default behavior: if Cobra can't find the subcommand,
	// run it as a script. Find the first non-flag argument to determine
	// the subcommand, so flags like --dry-run can appear before it.
	knownCmds := map[string]bool{
		"install": true, "i": true, "add": true, "remove": true,
		"help": true, "completion": true, "version": true,
		"_play-sound": true, "_play-music": true,
	}

	if scriptName := firstNonFlagArg(os.Args[1:]); scriptName != "" && !knownCmds[scriptName] {
		script := scriptName
		rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// args from Cobra contains positional args after flag parsing;
			// the first element is the script name, the rest are extra args.
			var extra []string
			if len(args) > 1 {
				extra = args[1:]
			}
			return run(script, extra)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// firstNonFlagArg returns the first argument that doesn't start with "-".
func firstNonFlagArg(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
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
	return runner.Run(args, dryRun, vibes && command == "install", notify)
}
