package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/progress"
	"github.com/decampsrenan/spm/internal/prompt"
	"github.com/decampsrenan/spm/internal/resolver"
	"github.com/decampsrenan/spm/internal/runner"
	"github.com/decampsrenan/spm/internal/ui"
)

var initCmd = &cobra.Command{
	Use:                "init [npm|yarn|pnpm|bun]",
	Short:              "Initialize a new project with a package manager",
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manual flag parsing since DisableFlagParsing is true
		// (needed to pass unknown flags like --react through to the PM).
		var filtered []string
		for _, a := range args {
			if a == "--dry-run" {
				dryRun = true
			} else if a == "--help" || a == "-h" {
				return cmd.Help()
			} else {
				filtered = append(filtered, a)
			}
		}
		return runInit(filtered)
	},
}

var validPMs = map[string]detector.PackageManager{
	"npm":  detector.NPM,
	"yarn": detector.Yarn,
	"pnpm": detector.Pnpm,
	"bun":  detector.Bun,
}

func runInit(args []string) error {
	// Check if package.json already exists
	if _, err := os.Stat("package.json"); err == nil {
		cwd, _ := os.Getwd()
		return fmt.Errorf("package.json already exists in %s", cwd)
	}

	// Determine PM and extra args
	var pm detector.PackageManager
	var extraArgs []string

	if len(args) > 0 {
		if parsed, ok := validPMs[args[0]]; ok {
			pm = parsed
			extraArgs = args[1:]
		} else {
			return fmt.Errorf("unknown package manager %q (valid: npm, yarn, pnpm, bun)", args[0])
		}
	} else {
		selected, err := prompt.SelectPM()
		if err != nil {
			return err
		}
		pm = selected
	}

	isTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	// Run <pm> init
	initArgs := resolver.Resolve(pm, "init", extraArgs)
	if isTTY && !rawOutput {
		if err := progress.Run(progress.Config{
			Args:   initArgs,
			DryRun: dryRun,
			Action: "Initializing",
			Done:   "Initialized",
		}); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}
	} else {
		if err := runner.RunSubprocess(initArgs, dryRun); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}
	}

	// Run <pm> install to generate lock file
	installArgs := resolver.Resolve(pm, "install", nil)
	if isTTY && !rawOutput {
		if err := progress.Run(progress.Config{
			Args:   installArgs,
			DryRun: dryRun,
			Action: "Installing dependencies",
			Done:   "Dependencies installed",
		}); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	} else {
		if err := runner.RunSubprocess(installArgs, dryRun); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}

	if !dryRun {
		ui.Println(ui.Success("Project initialized with " + string(pm)))
	}

	return nil
}
