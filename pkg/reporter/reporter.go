package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rafael/owl/cli/pkg/browser"
	"github.com/rafael/owl/cli/pkg/monitor"
)

// ReportFormat represents the format of the test report
type ReportFormat string

const (
	FormatHTML ReportFormat = "html"
	FormatJSON ReportFormat = "json"
)

// ReportType represents the type of test report
type ReportType string

const (
	ReportTypeHTTP    ReportType = "http"
	ReportTypeBrowser ReportType = "browser"
)

// GenerateReport generates a report for test results in the specified format
func GenerateReport(templatePath string, resultsDir string, reportFormat ReportFormat, reportType ReportType, httpResults []monitor.Result, browserResults []*browser.BrowserResult) (string, error) {
	// Create results directory if it doesn't exist
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create results directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.%s", reportType, timestamp, reportFormat)
	reportPath := filepath.Join(resultsDir, filename)

	switch reportFormat {
	case FormatHTML:
		return generateHTMLReport(templatePath, reportPath, reportType, httpResults, browserResults)
	case FormatJSON:
		return generateJSONReport(reportPath, reportType, httpResults, browserResults)
	default:
		return "", fmt.Errorf("unknown report format: %s", reportFormat)
	}
}

// generateHTMLReport generates an HTML report for test results
func generateHTMLReport(templatePath string, reportPath string, reportType ReportType, httpResults []monitor.Result, browserResults []*browser.BrowserResult) (string, error) {
	// Read template
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Generate report content based on type
	var reportContent string
	switch reportType {
	case ReportTypeHTTP:
		reportContent = generateHTTPHTMLReport(string(template), httpResults)
	case ReportTypeBrowser:
		reportContent = generateBrowserHTMLReport(string(template), browserResults)
	default:
		return "", fmt.Errorf("unknown report type: %s", reportType)
	}

	// Write report
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return reportPath, nil
}

// generateJSONReport generates a JSON report for test results
func generateJSONReport(reportPath string, reportType ReportType, httpResults []monitor.Result, browserResults []*browser.BrowserResult) (string, error) {
	var reportContent []byte
	var err error

	switch reportType {
	case ReportTypeHTTP:
		reportContent, err = generateHTTPJSONReport(httpResults)
	case ReportTypeBrowser:
		reportContent, err = generateBrowserJSONReport(browserResults)
	default:
		return "", fmt.Errorf("unknown report type: %s", reportType)
	}

	if err != nil {
		return "", err
	}

	// Write report
	if err := os.WriteFile(reportPath, reportContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return reportPath, nil
}

// JSONReport represents the structure of a JSON report
type JSONReport struct {
	ExecutedAt string         `json:"executed_at"`
	Type       string         `json:"type"`
	Summary    SummaryMetrics `json:"summary"`
	Tests      []JSONTestItem `json:"tests"`
}

// SummaryMetrics holds summary statistics
type SummaryMetrics struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// JSONTestItem represents a single test in JSON format
type JSONTestItem struct {
	ID       string `json:"id"`
	Suite    string `json:"suite"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// generateHTTPJSONReport generates JSON report for HTTP tests
func generateHTTPJSONReport(results []monitor.Result) ([]byte, error) {
	total := len(results)
	passed := 0
	failed := 0

	var tests []JSONTestItem
	for i, r := range results {
		status := "pass"
		if !r.Pass {
			status = "fail"
			failed++
		} else {
			passed++
		}

		errorMsg := ""
		if r.Error != nil {
			errorMsg = r.Error.Error()
		}

		tests = append(tests, JSONTestItem{
			ID:       fmt.Sprintf("HTTP-%d", i+1),
			Suite:    "HTTP Tests",
			Name:     r.Name,
			Duration: r.ResponseTime.String(),
			Status:   status,
			Error:    errorMsg,
		})
	}

	report := JSONReport{
		ExecutedAt: time.Now().Format(time.RFC3339),
		Type:       "http",
		Summary: SummaryMetrics{
			Total:   total,
			Passed:  passed,
			Failed:  failed,
			Skipped: 0,
		},
		Tests: tests,
	}

	return json.MarshalIndent(report, "", "  ")
}

// generateBrowserJSONReport generates JSON report for browser tests
func generateBrowserJSONReport(results []*browser.BrowserResult) ([]byte, error) {
	total := len(results)
	passed := 0
	failed := 0

	var tests []JSONTestItem
	for i, r := range results {
		status := "pass"
		if !r.Pass {
			status = "fail"
			failed++
		} else {
			passed++
		}

		errorMsg := ""
		if r.Error != nil {
			errorMsg = r.Error.Error()
		}

		tests = append(tests, JSONTestItem{
			ID:       fmt.Sprintf("BRW-%d", i+1),
			Suite:    "Browser Tests",
			Name:     r.Name,
			Duration: "-",
			Status:   status,
			Error:    errorMsg,
		})
	}

	report := JSONReport{
		ExecutedAt: time.Now().Format(time.RFC3339),
		Type:       "browser",
		Summary: SummaryMetrics{
			Total:   total,
			Passed:  passed,
			Failed:  failed,
			Skipped: 0,
		},
		Tests: tests,
	}

	return json.MarshalIndent(report, "", "  ")
}

// TestReportItem represents a single test result for the HTML report
type TestReportItem struct {
	ID       string
	Suite    string
	Name     string
	Duration string
	Status   string
	Error    string
}

// generateHTTPHTMLReport generates the HTML report for HTTP tests
func generateHTTPHTMLReport(template string, results []monitor.Result) string {
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

// generateBrowserHTMLReport generates the HTML report for browser tests
func generateBrowserHTMLReport(template string, results []*browser.BrowserResult) string {
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
