package cmd

import (
	"testing"
)

func TestUpgradeCmdFlags(t *testing.T) {
	if upgradeCmd.Flags().Lookup("alpha") == nil {
		t.Fatal("expected --alpha flag to be defined")
	}
	if upgradeCmd.Flags().Lookup("force") == nil {
		t.Fatal("expected --force flag to be defined")
	}
}

func TestRunUpgradeDevVersion(t *testing.T) {
	old := rootCmd.Version
	rootCmd.Version = "dev"
	t.Cleanup(func() { rootCmd.Version = old })

	err := runUpgrade()
	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}
