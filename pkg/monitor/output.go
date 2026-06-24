package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Color output helpers
var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
)

// PrintResult prints a test result in a formatted way
func PrintResult(result Result) {
	bold.Printf("\n%s\n", result.Name)
	fmt.Println(strings.Repeat("=", len(result.Name)))

	if result.Error != nil {
		red.Printf("  ✗ ERROR: %v\n", result.Error)
		return
	}

	// Status badge
	statusBadge := fmt.Sprintf("[%d] %s", result.StatusCode, result.ResponseTime.Round(time.Millisecond))
	if result.Pass {
		green.Printf("  ✓ PASS %s\n", statusBadge)
	} else {
		red.Printf("  ✗ FAIL %s\n", statusBadge)
	}

	// Assertions
	for _, a := range result.Assertions {
		if a.Pass {
			green.Printf("    ✓ %s: %s\n", a.Type, a.Message)
		} else {
			red.Printf("    ✗ %s: %s\n", a.Type, a.Message)
		}
	}
}

// PrintSummary prints a summary of all test results
func PrintSummary(results []Result) {
	passed := 0
	failed := 0

	for _, r := range results {
		if r.Pass {
			passed++
		} else if r.Error == nil {
			failed++
		}
	}

	fmt.Println()
	bold.Println("SUMMARY")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Total:  %d\n", len(results))
	green.Printf("Passed: %d\n", passed)
	if failed > 0 {
		red.Printf("Failed: %d\n", failed)
	} else {
		fmt.Printf("Failed: %d\n", failed)
	}
}