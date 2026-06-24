package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "owl is a high-performance, concurrent API monitoring CLI",
	Long:  `owl is a high-performance, concurrent API monitoring CLI application.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
}