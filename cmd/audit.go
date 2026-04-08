package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/decampsrenan/spm/internal/audio"
	"github.com/decampsrenan/spm/internal/audit"
	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/ecosystem"
	"github.com/decampsrenan/spm/internal/prompt"
)

var auditProdOnly bool
var auditJSON bool
var auditSeverity string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run a security audit on dependencies",
	Long:  "Runs the package manager's audit command, normalizes the output across npm/yarn/pnpm.",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		eco := ecosystem.ForPM(det.PM)
		if eco == nil {
			return fmt.Errorf("unsupported package manager: %s", det.PM)
		}

		opts := audit.Options{
			ProdOnly: auditProdOnly,
			JSON:     auditJSON,
			Notify:   notify,
			DryRun:   dryRun,
		}

		if auditSeverity != "" {
			sev, ok := audit.ParseSeverity(auditSeverity)
			if !ok {
				return fmt.Errorf("invalid severity %q (valid: info, low, moderate, high, critical)", auditSeverity)
			}
			opts.Severity = sev
		}

		exitCode, err := audit.Run(eco, det.Dir, opts)
		if err != nil {
			if notify {
				_ = audio.PlayNotification(audio.SoundError)
			}
			return err
		}

		if notify {
			if exitCode == audit.ExitClean {
				_ = audio.PlayNotification(audio.SoundSuccess)
			} else {
				_ = audio.PlayNotification(audio.SoundError)
			}
		}

		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	},
}

func init() {
	auditCmd.Flags().BoolVar(&auditProdOnly, "prod-only", false, "Only audit production dependencies")
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "Output results as JSON")
	auditCmd.Flags().StringVar(&auditSeverity, "severity", "", "Minimum severity to report (info, low, moderate, high, critical)")
}
