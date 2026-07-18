package cmd

import (
	"fmt"
	"os"

	"github.com/rafael/owl/cli/pkg/browser"
	"github.com/rafael/owl/cli/pkg/monitor"
	"github.com/spf13/cobra"
)

var verbose bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests from YAML files",
	Long: `Executes tests from YAML files (HTTP or browser tests).
Accepts a single argument: a .yaml/.yml file path or a directory containing such files.

The test type is automatically detected based on the YAML structure:
- HTTP tests: have a 'request:' field
- Browser tests: have a 'browser:' field`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat path: %w", err)
		}

		// If it's a directory, run both HTTP and browser tests found in it
		if info.IsDir() {
			return runAllTestsInDirectory(path)
		}

		// For a single file, detect the test type
		testType, err := detectTestType(path)
		if err != nil {
			return fmt.Errorf("failed to detect test type: %w", err)
		}

		switch testType {
		case "http":
			return runHTTPTests(path)
		case "browser":
			return runBrowserTests(path)
		default:
			return fmt.Errorf("unknown test type: %s", testType)
		}
	},
}

func init() {
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output on failure")
}

// detectTestType peeks at the first bytes of a YAML file to detect if it's HTTP or browser test
func detectTestType(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a directory, look for files inside
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}

		hasBrowser := false
		hasHTTP := false

		for _, entry := range entries {
			if entry.IsDir() {
				// Recursively check subdirectories
				subType, err := detectTestType(path + "/" + entry.Name())
				if err == nil {
					if subType == "browser" {
						hasBrowser = true
					} else if subType == "http" {
						hasHTTP = true
					}
				}
				continue
			}

			ext := entry.Name()
			if len(ext) < 5 {
				continue
			}
			ext = ext[len(ext)-5:]
			if ext != ".yaml" && ext != ".yml" {
				continue
			}

			filePath := path + "/" + entry.Name()
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			content := string(data)
			if containsSubstring(content, "browser:") {
				hasBrowser = true
			}
			if containsSubstring(content, "request:") {
				hasHTTP = true
			}
		}

		// If directory has both types, prefer HTTP for backward compatibility
		// If only one type, return that type
		if hasBrowser && !hasHTTP {
			return "browser", nil
		}
		// Default to HTTP for backward compatibility
		return "http", nil
	}

	// It's a file - read and detect
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Simple detection based on YAML content
	content := string(data)

	if containsSubstring(content, "browser:") {
		return "browser", nil
	}

	if containsSubstring(content, "request:") {
		return "http", nil
	}

	// Default to HTTP for backwards compatibility
	return "http", nil
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func runHTTPTests(path string) error {
	// Find all test files
	files, err := monitor.FindTestFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No .yaml or .yml files found.")
		return nil
	}

	fmt.Printf("Found %d HTTP test file(s)\n", len(files))

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
		monitor.PrintResult(result, verbose)
	}

	monitor.PrintSummary(results)

	// Exit with error if any test failed
	for _, r := range results {
		if !r.Pass {
			os.Exit(1)
		}
	}

	return nil
}

// runAllTestsInDirectory finds and runs both HTTP and browser tests in a directory
func runAllTestsInDirectory(path string) error {
	httpPass := true
	browserPass := true

	// Run HTTP tests
	httpFiles, err := monitor.FindTestFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find HTTP test files: %w", err)
	}

	if len(httpFiles) > 0 {
		fmt.Printf("Found %d HTTP test file(s)\n", len(httpFiles))
		results := make([]monitor.Result, 0, len(httpFiles))
		for _, file := range httpFiles {
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
			monitor.PrintResult(result, verbose)
		}
		monitor.PrintSummary(results)
		for _, r := range results {
			if !r.Pass {
				httpPass = false
			}
		}
	}

	// Run browser tests
	browserFiles, err := browser.FindBrowserTestFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find browser test files: %w", err)
	}

	if len(browserFiles) > 0 {
		fmt.Printf("Found %d browser test file(s)\n", len(browserFiles))
		var results []*browser.BrowserResult
		for _, file := range browserFiles {
			config, err := browser.LoadBrowserTest(file)
			if err != nil {
				results = append(results, &browser.BrowserResult{
					Name:  file,
					Pass:  false,
					Error: fmt.Errorf("failed to load: %w", err),
				})
				continue
			}
			result := browser.ExecuteBrowserTest(*config)
			results = append(results, result)
			browser.PrintBrowserResult(result, verbose)
		}
		browser.PrintBrowserSummary(results)
		for _, r := range results {
			if !r.Pass {
				browserPass = false
			}
		}
	}

	if len(httpFiles) == 0 && len(browserFiles) == 0 {
		fmt.Println("No .yaml or .yml files found.")
		return nil
	}

	// Exit with error if any test failed
	if !httpPass || !browserPass {
		os.Exit(1)
	}

	return nil
}

func runBrowserTests(path string) error {
	// Find all test files
	files, err := browser.FindBrowserTestFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find test files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No .yaml or .yml files found.")
		return nil
	}

	fmt.Printf("Found %d browser test file(s)\n", len(files))

	// Execute each test
	var results []*browser.BrowserResult
	for _, file := range files {
		config, err := browser.LoadBrowserTest(file)
		if err != nil {
			results = append(results, &browser.BrowserResult{
				Name:  file,
				Pass:  false,
				Error: fmt.Errorf("failed to load: %w", err),
			})
			continue
		}

		result := browser.ExecuteBrowserTest(*config)
		results = append(results, result)
		browser.PrintBrowserResult(result, verbose)
	}

	browser.PrintBrowserSummary(results)

	// Exit with error if any test failed
	for _, r := range results {
		if !r.Pass {
			os.Exit(1)
		}
	}

	return nil
}