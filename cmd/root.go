package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/prompt"
	"github.com/decampsrenan/spm/internal/resolver"
	"github.com/decampsrenan/spm/internal/runner"
	"github.com/decampsrenan/spm/internal/scripts"
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
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify at least one package to add\n\nUsage: spm add <package> [packages...]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("add", args)
	},
}

var runCmd = &cobra.Command{
	Use:   "run [script]",
	Short: "Run a script from package.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return run(args[0], args[1:])
		}

		// No script specified — show interactive selection
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot get working directory: %w", err)
		}

		det, err := detect(cwd)
		if err != nil {
			return err
		}

		scriptList, err := scripts.List(det.Dir)
		if err != nil {
			return err
		}
		if len(scriptList) == 0 {
			return fmt.Errorf("no scripts found in package.json")
		}

		names := make([]string, len(scriptList))
		cmds := make([]string, len(scriptList))
		for i, s := range scriptList {
			names[i] = s.Name
			cmds[i] = s.Command
		}

		selected, err := prompt.SelectScript(names, cmds)
		if err != nil {
			return err
		}

		resolved := resolver.Resolve(det.PM, selected, nil)
		return runner.Run(resolved, dryRun, false, notify)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [packages...]",
	Short: "Remove one or more packages",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify at least one package to remove\n\nUsage: spm remove <package> [packages...]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run("remove", args)
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove node_modules and optionally the lock file",
	RunE: func(cmd *cobra.Command, args []string) error {
		lock, _ := cmd.Flags().GetBool("lock")
		yes, _ := cmd.Flags().GetBool("yes")
		return runClean(lock, yes)
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
	installCmd.FParseErrWhitelist.UnknownFlags = true
	addCmd.FParseErrWhitelist.UnknownFlags = true
	runCmd.FParseErrWhitelist.UnknownFlags = true
	removeCmd.FParseErrWhitelist.UnknownFlags = true
	cleanCmd.Flags().Bool("lock", false, "Also remove the lock file")
	cleanCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(cleanCmd)
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
		"install": true, "i": true, "add": true, "run": true, "remove": true, "clean": true,
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

func runClean(lock bool, yes bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}

	var det detector.Detection

	if lock {
		// Need PM to determine lock file name — use detect() which handles
		// ErrNoLockFile by prompting and multiple detections by selecting.
		det, err = detect(cwd)
		if err != nil {
			return err
		}
	} else {
		// Only need the project dir for node_modules removal.
		detections, err := detector.Detect(cwd)
		var noLock *detector.ErrNoLockFile
		if errors.As(err, &noLock) {
			det.Dir = noLock.Dir
		} else if err != nil {
			return err
		} else if len(detections) == 1 {
			det = detections[0]
		} else {
			det, err = prompt.Select(detections)
			if err != nil {
				return err
			}
		}
	}

	targets := []string{"node_modules"}
	if lock {
		lockFile := detector.LockFileName(det.PM)
		if lockFile != "" {
			targets = append(targets, lockFile)
		}
	}

	// Filter to targets that actually exist on disk.
	var existing []string
	for _, t := range targets {
		path := filepath.Join(det.Dir, t)
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, t)
		}
	}

	if len(existing) == 0 {
		fmt.Println("Nothing to remove.")
		return nil
	}

	fmt.Println("The following will be removed:")
	for _, t := range existing {
		fmt.Printf("  %s\n", filepath.Join(det.Dir, t))
	}

	if dryRun {
		fmt.Println("(dry-run: nothing was deleted)")
		return nil
	}

	if !yes {
		confirmed, err := prompt.Confirm("Proceed?")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	for _, t := range existing {
		path := filepath.Join(det.Dir, t)
		if t == "node_modules" {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
		fmt.Printf("Removed %s\n", path)
	}

	return nil
}

func run(command string, extraArgs []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get working directory: %w", err)
	}

	det, err := detect(cwd)
	if err != nil {
		return err
	}

	args := resolver.Resolve(det.PM, command, extraArgs)
	return runner.Run(args, dryRun, vibes && command == "install", notify)
}

// detect finds the package manager for the project rooted at cwd.
// If no lock file exists, it prompts the user to choose one.
func detect(cwd string) (detector.Detection, error) {
	detections, err := detector.Detect(cwd)

	var noLock *detector.ErrNoLockFile
	if errors.As(err, &noLock) {
		return prompt.SelectFromAll(noLock.Dir)
	}
	if err != nil {
		return detector.Detection{}, err
	}

	if len(detections) == 1 {
		return detections[0], nil
	}

	return prompt.Select(detections)
}
