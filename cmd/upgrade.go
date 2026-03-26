package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/ui"
	"github.com/decampsrenan/spm/internal/updater"
)

var upgradeAlpha bool
var upgradeForce bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade spm to the latest version",
	Long:  "Downloads and installs the latest spm release from GitHub. Use --alpha to include pre-releases.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpgrade()
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeAlpha, "alpha", false, "Include alpha/pre-release versions")
	upgradeCmd.Flags().BoolVar(&upgradeForce, "force", false, "Reinstall even if already up to date")
}

func runUpgrade() error {
	fetcher := &updater.GitHubFetcher{Repo: "decampsrenan/spm"}
	opts := updater.Options{
		CurrentVersion: rootCmd.Version,
		Alpha:          upgradeAlpha,
		Force:          upgradeForce,
	}

	result, err := updater.Plan(fetcher, opts)
	if err != nil {
		return err
	}

	if result.AlreadyLatest {
		ui.Println(ui.Success(fmt.Sprintf("Already up to date (%s)", result.CurrentVersion)))
		ui.Println(ui.Dim("Use --force to reinstall anyway."))
		return nil
	}

	ui.Println(ui.Info(fmt.Sprintf("%s → %s", result.CurrentVersion, result.LatestVersion)))
	ui.Println(ui.Dim(fmt.Sprintf("Target: %s", result.TargetPath)))

	if dryRun {
		ui.Println(ui.Dim(fmt.Sprintf("Would download: %s", result.DownloadURL)))
		ui.Println(ui.Dim("(dry-run: nothing was downloaded)"))
		return nil
	}

	ui.Println(ui.Dim(fmt.Sprintf("Downloading %s...", result.LatestVersion)))

	if err := updater.Execute(result); err != nil {
		return err
	}

	ui.Println(ui.Success(fmt.Sprintf("Upgraded spm to %s", result.LatestVersion)))
	return nil
}
