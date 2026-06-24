package cmd

import (
	"fmt"
	"os"

	"github.com/rafael/owl/cli/pkg/monitor"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the API monitoring task",
	Long: `Executes the API monitoring task with concurrent request handling.
Accepts a single argument: a .yaml/.yml file path or a directory containing such files.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		// Find all test files
		files, err := monitor.FindTestFiles(path)
		if err != nil {
			return fmt.Errorf("failed to find test files: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No .yaml or .yml files found.")
			return nil
		}

		fmt.Printf("Found %d test file(s)\n", len(files))

		// Execute each test
		results := make([]monitor.Result, 0, len(files))
		for _, file := range files {
			config, err := monitor.LoadTest(file)
			if err != nil {
				results = append(results, monitor.Result{
					Name:  file,
					Pass:  false,
					Error: fmt.Errorf("failed to load: %w", err),
				})
				continue
			}

			result := monitor.Execute(config)
			results = append(results, result)
			monitor.PrintResult(result)
		}

		monitor.PrintSummary(results)

		// Exit with error if any test failed
		for _, r := range results {
			if !r.Pass {
				os.Exit(1)
			}
		}

		return nil
	},
}