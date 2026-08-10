package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/rafael/owl/cli/pkg/config"
	"github.com/rafael/owl/cli/pkg/script"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// BrowserTestConfig represents the full parsed YAML test structure
type BrowserTestConfig struct {
	Version        int                       `yaml:"version"`
	Metadata       MetadataConfig            `yaml:"metadata"`
	Browser        BrowserConfig             `yaml:"browser"`
	Steps          []StepConfig              `yaml:"steps"`
	Assertions     []BrowserAssertionConfig  `yaml:"assertions"`
	Retries        int                      `yaml:"retries"`
	RetryInterval  int                      `yaml:"retry_interval"`
	BeforeScript   string                   `yaml:"before_script"`
	AfterScript    string                   `yaml:"after_script"`
	ScriptTimeout  int                      `yaml:"script_timeout"`
	WorkDir        string                   `yaml:"-"`
}

// MetadataConfig holds test metadata
type MetadataConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

// BrowserConfig holds browser launch options
type BrowserConfig struct {
	URL            string         `yaml:"url"`
	Headless       bool           `yaml:"headless"`
	TimeoutSeconds int            `yaml:"timeout_seconds"`
	WaitUntil      string         `yaml:"wait_until"`
	Viewport       ViewportConfig `yaml:"viewport"`
	Proxy          string         `yaml:"proxy"`
	UserAgent      string         `yaml:"user_agent"`
}

// ViewportConfig holds browser viewport dimensions
type ViewportConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// StepConfig represents a single browser action step
type StepConfig struct {
	Name           string      `yaml:"name"`
	Action         string      `yaml:"action"`
	Selector       string      `yaml:"selector"`
	Value          interface{} `yaml:"value"`
	URL            string      `yaml:"url"`
	Target         string      `yaml:"target"`
	Filter         string      `yaml:"filter"`
	Pattern        string      `yaml:"pattern"`
	Script         string      `yaml:"script"`
	TimeoutSeconds int         `yaml:"timeout_seconds"`
	WaitMillis     int         `yaml:"wait"`
	Path           string      `yaml:"path"`
	Key            string      `yaml:"key"`
	Width          int         `yaml:"width"`
	Height         int         `yaml:"height"`
}

// BrowserAssertionConfig represents a browser assertion
type BrowserAssertionConfig struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Selector string      `yaml:"selector"`
	Path     string      `yaml:"path"`
	Pattern  string      `yaml:"pattern"`
	Key      string      `yaml:"key"`
	Expected interface{} `yaml:"expected"`
	Level    string      `yaml:"level"`
	Script   string      `yaml:"script"`
}

// BrowserResult holds the test execution result
type BrowserResult struct {
	Name           string
	Pass           bool
	Steps          []StepResult
	Assertions     []AssertionResult
	Error          error
	ConsoleLogs    []ConsoleLog
	ScreenshotPath string
}

// StepResult holds the result of a single step execution
type StepResult struct {
	Name   string
	Pass   bool
	Error  error
	Output string
}

// AssertionResult holds the result of a single assertion
type AssertionResult struct {
	Name   string
	Pass   bool
	Error  error
	Actual interface{}
}

// ConsoleLog represents a browser console log entry
type ConsoleLog struct {
	Level string
	Text  string
}

// ExecuteBrowserTest runs a browser test based on the provided configuration
func ExecuteBrowserTest(config BrowserTestConfig, visibleOverride ...bool) *BrowserResult {
	return ExecuteBrowserTestWithRetry(config, config.Retries, config.RetryInterval, visibleOverride...)
}

// ExecuteBrowserTestWithRetry runs a browser test with retry support
func ExecuteBrowserTestWithRetry(config BrowserTestConfig, retryCount int, retryIntervalMs int, visibleOverride ...bool) *BrowserResult {
	// Set defaults
	if config.Browser.TimeoutSeconds == 0 {
		config.Browser.TimeoutSeconds = 30
	}
	if config.Browser.WaitUntil == "" {
		config.Browser.WaitUntil = "networkidle2"
	}
	if config.Browser.Viewport.Width == 0 {
		config.Browser.Viewport.Width = 1280
	}
	if config.Browser.Viewport.Height == 0 {
		config.Browser.Viewport.Height = 720
	}

	// Determine if browser should be headless
	isHeadless := config.Browser.Headless
	if len(visibleOverride) > 0 && visibleOverride[0] {
		isHeadless = false
	}

	// Set default script timeout if not specified
	scriptTimeout := 30 * time.Second
	if config.ScriptTimeout > 0 {
		scriptTimeout = time.Duration(config.ScriptTimeout) * time.Second
	}

	// Run before_script if defined (before any attempts)
	if config.BeforeScript != "" {
		runner := script.NewRunner(scriptTimeout)
		if err := runner.Run(config.BeforeScript, config.WorkDir); err != nil {
			return &BrowserResult{
				Name:       config.Metadata.Name,
				Pass:       false,
				Steps:      []StepResult{},
				Assertions: []AssertionResult{},
				Error:      fmt.Errorf("before_script failed: %w", err),
			}
		}
	}

	retryInterval := time.Duration(retryIntervalMs) * time.Millisecond

	var lastErr error
	var testResult *BrowserResult
	for attempt := 0; attempt <= retryCount; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(retryInterval)
		}

		result := executeBrowserTestAttempt(config, isHeadless)
		testResult = result

		// Check if we should retry
		if result.Pass {
			// Success - break and run after_script
			break
		}

		// If no error but test failed, it means an assertion failed - don't retry
		if result.Error == nil {
			break
		}

		// Check if it's a retryable error (timeout, network, etc.)
		if isRetryableBrowserError(result.Error) {
			lastErr = result.Error
			continue
		}

		// Non-retryable error (assertion error)
		break

		lastErr = result.Error
	}

	// All retries exhausted
	if testResult == nil || (testResult != nil && !testResult.Pass && testResult.Error == nil && lastErr != nil) {
		testResult = &BrowserResult{
			Name:       config.Metadata.Name,
			Pass:       false,
			Steps:      []StepResult{},
			Assertions: []AssertionResult{},
			Error:      fmt.Errorf("all %d retry attempts failed: %w", retryCount, lastErr),
		}
	}

	// Run after_script if defined (even if test failed)
	if config.AfterScript != "" {
		runner := script.NewRunner(scriptTimeout)
		if err := runner.Run(config.AfterScript, config.WorkDir); err != nil {
			// Log the error but don't change the test result
			if testResult.Error == nil {
				testResult.Error = fmt.Errorf("after_script failed: %w", err)
			}
		}
	}

	return testResult
}

// executeBrowserTestAttempt performs a single browser test attempt
func executeBrowserTestAttempt(config BrowserTestConfig, isHeadless bool) *BrowserResult {
	result := &BrowserResult{
		Name:       config.Metadata.Name,
		Pass:       true,
		Steps:      make([]StepResult, 0),
		Assertions: make([]AssertionResult, 0),
		ConsoleLogs: make([]ConsoleLog, 0),
	}

	// Launch browser
	browserLauncher := launcher.New().
		Headless(isHeadless).
		Set("viewport", fmt.Sprintf("%d,%d", config.Browser.Viewport.Width, config.Browser.Viewport.Height)).
		Set("no-sandbox", "").
		Set("disable-setuid-sandbox", "")

	if config.Browser.UserAgent != "" {
		browserLauncher.Set("user-agent", config.Browser.UserAgent)
	}

	if config.Browser.Proxy != "" {
		browserLauncher.Proxy(config.Browser.Proxy)
	}

	browserURL := browserLauncher.MustLaunch()

	browser := rod.New().ControlURL(browserURL)
	defer browser.Close()

	err := browser.Connect()
	if err != nil {
		result.Pass = false
		result.Error = fmt.Errorf("failed to connect to browser: %w", err)
		return result
	}

	// Create a timeout context - use a large default if not specified
	browserTimeout := time.Duration(config.Browser.TimeoutSeconds) * time.Second
	if browserTimeout == 0 {
		browserTimeout = 120 * time.Second // Default 120 seconds
	}
	ctx, cancel := context.WithTimeout(context.Background(), browserTimeout)
	defer cancel()

	page := browser.MustPage()
	defer page.Close()

	// Navigate to initial URL if provided
	if config.Browser.URL != "" {
		err = page.Timeout(browserTimeout).Navigate(config.Browser.URL)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to navigate to initial URL: %w", err)
			return result
		}
	}

	// Execute steps
	for _, step := range config.Steps {
		stepResult := ExecuteStep(page, step, ctx)
		result.Steps = append(result.Steps, stepResult)
		if !stepResult.Pass && stepResult.Error != nil {
			result.Pass = false
		}
	}

	// Execute assertions
	for _, assertion := range config.Assertions {
		assertionResult := ValidateBrowserAssertion(page, assertion, ctx)
		result.Assertions = append(result.Assertions, assertionResult)
		if !assertionResult.Pass {
			result.Pass = false
		}
	}

	return result
}

// isRetryableBrowserError checks if an error should trigger a retry
func isRetryableBrowserError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"context deadline",
		"connection",
		"network",
		"i/o timeout",
		"temporary failure",
		"name or service not known",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// LoadBrowserTest loads a browser test from a YAML file
func LoadBrowserTest(path string) (*BrowserTestConfig, error) {
	return LoadBrowserTestWithConfig(path, nil)
}

// LoadBrowserTestWithConfig loads a browser test from a YAML file and resolves placeholders
func LoadBrowserTestWithConfig(path string, cfg *config.Config) (*BrowserTestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// Resolve placeholders if config is provided
	if cfg != nil {
		content, err = cfg.ResolveValues(content)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve config: %w", err)
		}
	}

	var config BrowserTestConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate version
	if config.Version != 1 {
		return nil, fmt.Errorf("unsupported version: %d (supported: 1)", config.Version)
	}

	// Validate required fields
	if config.Metadata.Name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}

	// Set the working directory to the test file's directory
	config.WorkDir = filepath.Dir(path)

	return &config, nil
}

// FindBrowserTestFiles finds all .yaml and .yml files in a directory that are browser tests
func FindBrowserTestFiles(path string) ([]string, error) {
	var files []string

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subFiles, err := FindBrowserTestFiles(filepath.Join(path, entry.Name()))
				if err != nil {
					continue
				}
				files = append(files, subFiles...)
			} else {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".yaml" || ext == ".yml" {
					filePath := filepath.Join(path, entry.Name())
					// Only include files that contain "browser:" (browser test)
					if isBrowserTestFile(filePath) {
						files = append(files, filePath)
					}
				}
			}
		}
	} else {
		// For a single file, only include if it's a browser test
		if isBrowserTestFile(path) {
			files = append(files, path)
		}
	}

	return files, nil
}

// isBrowserTestFile checks if a file is a browser test by looking for "browser:" field
func isBrowserTestFile(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, "browser:")
}

// NewBrowserCommand creates the 'owl browser' command
func NewBrowserCommand() *cobra.Command {
	browserCmd := &cobra.Command{
		Use:   "browser",
		Short: "Browser automation tests using go-rod",
		Long:  `Run browser automation tests using the go-rod library for Chrome/Edge automation`,
	}

	browserCmd.AddCommand(&cobra.Command{
		Use:   "run [file or directory]",
		Short: "Run browser tests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			verbose, _ := cmd.Flags().GetBool("verbose")

			files, err := FindBrowserTestFiles(path)
			if err != nil {
				return fmt.Errorf("failed to find test files: %w", err)
			}

			if len(files) == 0 {
				fmt.Println("No browser test files found")
				return nil
			}

			var results []*BrowserResult
			for _, file := range files {
				config, err := LoadBrowserTest(file)
				if err != nil {
					fmt.Printf("Failed to load %s: %v\n", file, err)
					continue
				}

				result := ExecuteBrowserTest(*config)
				results = append(results, result)

				PrintBrowserResult(result, verbose)

				if !result.Pass {
					screenshotPath, _ := cmd.Flags().GetString("screenshot")
					if screenshotPath != "" {
						fmt.Printf("Screenshot would be saved to: %s\n", screenshotPath)
					}
				}
			}

			PrintBrowserSummary(results)

			// Return error if any test failed
			for _, r := range results {
				if !r.Pass {
					return fmt.Errorf("some tests failed")
				}
			}

			return nil
		},
	})

	browserCmd.PersistentFlags().BoolP("verbose", "v", false, "Show verbose output including screenshots and console logs")
	browserCmd.PersistentFlags().String("screenshot", "", "Path to save screenshot on failure")

	return browserCmd
}
