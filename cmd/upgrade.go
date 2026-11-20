package cmd

import (
	"fmt"

	"github.com/rafael/owl/cli/pkg/updater"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade owl to the latest version",
	Long:  `Checks for updates and upgrades owl to the latest available version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Checking for updates...\n")
		info, err := updater.CheckForUpdate(Version)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !info.IsUpdate {
			fmt.Printf("You are already running the latest version: %s\n", Version)
			return nil
		}

		fmt.Printf("A new version is available: %s (current: %s)\n", info.LatestVersion, Version)
		fmt.Printf("Release notes: %s\n", info.URL)

		if err := updater.Upgrade(Version); err != nil {
			return fmt.Errorf("upgrade failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
