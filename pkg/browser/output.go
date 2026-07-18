package browser

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// PrintBrowserResult prints the result of a single browser test
func PrintBrowserResult(result *BrowserResult, verbose bool) {
	// Test name header
	nameColor := color.FgGreen
	if !result.Pass {
		nameColor = color.FgRed
	}
	color.New(nameColor).Printf("✗ %s", result.Name)
	if result.Pass {
		color.New(color.FgGreen).Print(" PASS")
	} else {
		color.New(color.FgRed).Print(" FAIL")
	}
	fmt.Println()

	// Print error if any
	if result.Error != nil {
		color.New(color.FgRed).Printf("  Error: %v\n", result.Error)
	}

	// Print steps
	for _, step := range result.Steps {
		stepPrefix := "  ✓"
		stepColor := color.FgGreen
		if !step.Pass {
			stepPrefix = "  ✗"
			stepColor = color.FgRed
		}
		color.New(stepColor).Printf("%s %s", stepPrefix, step.Name)
		if verbose && step.Output != "" {
			fmt.Printf(" (%s)", step.Output)
		}
		fmt.Println()
		if !step.Pass && step.Error != nil {
			color.New(color.FgRed).Printf("    Error: %v\n", step.Error)
		}
	}

	// Print assertions
	for _, assertion := range result.Assertions {
		assertPrefix := "  ✓"
		assertColor := color.FgGreen
		if !assertion.Pass {
			assertPrefix = "  ✗"
			assertColor = color.FgRed
		}
		color.New(assertColor).Printf("%s %s", assertPrefix, assertion.Name)
		if !assertion.Pass && assertion.Error != nil {
			fmt.Printf(": %v", assertion.Error)
		}
		fmt.Println()
	}

	// Verbose output
	if verbose {
		// Print console logs
		if len(result.ConsoleLogs) > 0 {
			fmt.Println("  Console Logs:")
			for _, log := range result.ConsoleLogs {
				levelColor := color.FgWhite
				switch log.Level {
				case "error":
					levelColor = color.FgRed
				case "warn":
					levelColor = color.FgYellow
				case "info":
					levelColor = color.FgBlue
				}
				color.New(levelColor).Printf("    [%s] %s\n", strings.ToUpper(log.Level), log.Text)
			}
		}

		// Print screenshot path if any
		if result.ScreenshotPath != "" {
			fmt.Printf("  Screenshot: %s\n", result.ScreenshotPath)
		}
	}

	fmt.Println()
}

// PrintBrowserSummary prints a summary of all browser test results
func PrintBrowserSummary(results []*BrowserResult) {
	if len(results) == 0 {
		return
	}

	passed := 0
	failed := 0

	for _, r := range results {
		if r.Pass {
			passed++
		} else {
			failed++
		}
	}

	total := len(results)

	// Summary box
	fmt.Println(strings.Repeat("─", 50))
	summaryColor := color.FgGreen
	if failed > 0 {
		summaryColor = color.FgRed
	}
	color.New(summaryColor).Printf("Browser Test Summary: %d/%d passed", passed, total)
	if failed > 0 {
		color.New(color.FgRed).Printf(" (%d failed)", failed)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))

	// Exit with non-zero code if any test failed
	if failed > 0 {
		fmt.Println("Some browser tests failed.")
	}
}
