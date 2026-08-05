package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafael/owl/cli/pkg/browser"
	"github.com/rafael/owl/cli/pkg/monitor"
)

// ReportType represents the type of test report
type ReportType string

const (
	ReportTypeHTTP    ReportType = "http"
	ReportTypeBrowser ReportType = "browser"
)

// GenerateReport generates an HTML report for test results
func GenerateReport(templatePath string, resultsDir string, reportType ReportType, httpResults []monitor.Result, browserResults []*browser.BrowserResult) (string, error) {
	// Create results directory if it doesn't exist
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create results directory: %w", err)
	}

	// Read template
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Generate report content based on type
	var reportContent string
	switch reportType {
	case ReportTypeHTTP:
		reportContent = generateHTTPReport(string(template), httpResults)
	case ReportTypeBrowser:
		reportContent = generateBrowserReport(string(template), browserResults)
	default:
		return "", fmt.Errorf("unknown report type: %s", reportType)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.html", reportType, timestamp)
	reportPath := filepath.Join(resultsDir, filename)

	// Write report
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return reportPath, nil
}

// TestReportItem represents a single test result for the report
type TestReportItem struct {
	ID       string
	Suite    string
	Name     string
	Duration string
	Status   string
	Error    string
}

// generateHTTPReport generates the HTML report for HTTP tests
func generateHTTPReport(template string, results []monitor.Result) string {
	// Calculate metrics
	total := len(results)
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Pass {
			passed++
		} else {
			failed++
		}
	}

	// Generate table rows
	var rows []string
	for i, r := range results {
		status := "pass"
		if !r.Pass {
			status = "fail"
		}

		duration := r.ResponseTime.String()
		var errorMsg string
		if r.Error != nil {
			errorMsg = fmt.Sprintf("Error: %s", r.Error.Error())
		}

		item := TestReportItem{
			ID:       fmt.Sprintf("HTTP-%d", i+1),
			Suite:    "HTTP Tests",
			Name:     r.Name,
			Duration: duration,
			Status:   status,
			Error:    errorMsg,
		}

		rows = append(rows, formatTestRow(item))
	}

	// Replace placeholders in template
	report := template
	report = replacePlaceholder(report, "Executed on: 2026-08-05 12:00:00 | Environment: Production | Suite: Regression",
		fmt.Sprintf("Executed on: %s | Type: HTTP", time.Now().Format("2006-01-02 15:04:05")))
	report = replacePlaceholder(report, `<div class="card-value">3</div>`, fmt.Sprintf(`<div class="card-value">%d</div>`, total))
	report = replacePlaceholder(report, `<div class="card-value text-pass">2</div>`, fmt.Sprintf(`<div class="card-value text-pass">%d</div>`, passed))
	report = replacePlaceholder(report, `<div class="card-value text-fail">1</div>`, fmt.Sprintf(`<div class="card-value text-fail">%d</div>`, failed))
	report = replacePlaceholder(report, `<div class="card-value text-skip">0</div>`, `<div class="card-value text-skip">0</div>`)

	// Replace table body
	rowsStr := strings.Join(rows, "\n")
	report = replacePlaceholder(report, `<tr data-status="pass">
                    <td>TS-101</td>
                    <td>AuthTests</td>
                    <td>User login with valid credentials</td>
                    <td>1.2s</td>
                    <td><span class="badge badge-pass">Pass</span></td>
                </tr>
                <tr data-status="fail">
                    <td>TS-102</td>
                    <td>PaymentTests</td>
                    <td>Checkout process with expired credit card</td>
                    <td>3.5s</td>
                    <td>
                        <span class="badge badge-fail">Fail</span>
                        <div class="error-log">AssertionError: Expected status code 400 but got 500.<br> at PaymentTests.java:42</div>
                    </td>
                </tr>
                <tr data-status="pass">
                    <td>TS-103</td>
                    <td>ProfileTests</td>
                    <td>Update user profile avatar picture</td>
                    <td>0.8s</td>
                    <td><span class="badge badge-pass">Pass</span></td>
                </tr>`, rowsStr)

	return report
}

// generateBrowserReport generates the HTML report for browser tests
func generateBrowserReport(template string, results []*browser.BrowserResult) string {
	// Calculate metrics
	total := len(results)
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Pass {
			passed++
		} else {
			failed++
		}
	}

	// Generate table rows
	var rows []string
	for i, r := range results {
		status := "pass"
		if !r.Pass {
			status = "fail"
		}

		errorMsg := ""
		if r.Error != nil {
			errorMsg = fmt.Sprintf("Error: %s", r.Error.Error())
		}

		item := TestReportItem{
			ID:       fmt.Sprintf("BRW-%d", i+1),
			Suite:    "Browser Tests",
			Name:     r.Name,
			Duration: "-",
			Status:   status,
			Error:    errorMsg,
		}

		rows = append(rows, formatTestRow(item))
	}

	// Replace placeholders in template
	report := template
	report = replacePlaceholder(report, "Executed on: 2026-08-05 12:00:00 | Environment: Production | Suite: Regression",
		fmt.Sprintf("Executed on: %s | Type: Browser", time.Now().Format("2006-01-02 15:04:05")))
	report = replacePlaceholder(report, `<div class="card-value">3</div>`, fmt.Sprintf(`<div class="card-value">%d</div>`, total))
	report = replacePlaceholder(report, `<div class="card-value text-pass">2</div>`, fmt.Sprintf(`<div class="card-value text-pass">%d</div>`, passed))
	report = replacePlaceholder(report, `<div class="card-value text-fail">1</div>`, fmt.Sprintf(`<div class="card-value text-fail">%d</div>`, failed))
	report = replacePlaceholder(report, `<div class="card-value text-skip">0</div>`, `<div class="card-value text-skip">0</div>`)

	// Replace table body
	rowsStr := strings.Join(rows, "\n")
	report = replacePlaceholder(report, `<tr data-status="pass">
                    <td>TS-101</td>
                    <td>AuthTests</td>
                    <td>User login with valid credentials</td>
                    <td>1.2s</td>
                    <td><span class="badge badge-pass">Pass</span></td>
                </tr>
                <tr data-status="fail">
                    <td>TS-102</td>
                    <td>PaymentTests</td>
                    <td>Checkout process with expired credit card</td>
                    <td>3.5s</td>
                    <td>
                        <span class="badge badge-fail">Fail</span>
                        <div class="error-log">AssertionError: Expected status code 400 but got 500.<br> at PaymentTests.java:42</div>
                    </td>
                </tr>
                <tr data-status="pass">
                    <td>TS-103</td>
                    <td>ProfileTests</td>
                    <td>Update user profile avatar picture</td>
                    <td>0.8s</td>
                    <td><span class="badge badge-pass">Pass</span></td>
                </tr>`, rowsStr)

	return report
}

// formatTestRow formats a single test row for the HTML table
func formatTestRow(item TestReportItem) string {
	var statusBadge string
	var errorDiv string

	if item.Status == "pass" {
		statusBadge = `<span class="badge badge-pass">Pass</span>`
	} else {
		statusBadge = `<span class="badge badge-fail">Fail</span>`
		if item.Error != "" {
			errorDiv = fmt.Sprintf(`<div class="error-log">%s</div>`, item.Error)
		}
	}

	return fmt.Sprintf(`<tr data-status="%s">
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s%s</td>
                </tr>`, item.Status, item.ID, item.Suite, item.Name, item.Duration, statusBadge, errorDiv)
}

// replacePlaceholder replaces a placeholder string in the template
func replacePlaceholder(template, old, new string) string {
	return strings.Replace(template, old, new, 1)
}