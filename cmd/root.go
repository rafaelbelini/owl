package cmd

import (
	"fmt"
	"os"

	"github.com/rafael/owl/cli/pkg/updater"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "owl is a high-performance, concurrent API monitoring CLI",
	Long:  `owl is a high-performance, concurrent API monitoring CLI application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip version check if:
		// - --version flag was used
		// - running the upgrade command (to avoid infinite loop)
		// - OWL_NO_UPDATE_CHECK env var is set
		if cmd.Name() == "version" || cmd.Name() == "upgrade" {
			return
		}
		if os.Getenv("OWL_NO_UPDATE_CHECK") != "" {
			return
		}

		// Check for updates in background (non-blocking)
		go checkForUpdates()
	},
}

func checkForUpdates() {
	info, err := updater.CheckForUpdate(Version)
	if err != nil {
		return // Silently fail - update check should not block commands
	}

	if info.IsUpdate {
		yellow := "\033[33m"
		reset := "\033[0m"
		fmt.Printf("\n%s🐦 A new version is available: %s%s\n", yellow, info.LatestVersion, reset)
		fmt.Printf("  Current: %s  →  Latest: %s\n", Version, info.LatestVersion)
		fmt.Printf("  To upgrade, run: %sowl upgrade%s\n\n", yellow, reset)
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(upgradeCmd)

	// Add version
	rootCmd.Version = Version
}
