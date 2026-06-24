package monitor

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TestConfig represents the parsed YAML configuration
type TestConfig struct {
	Version        int                    `yaml:"version"`
	Metadata       MetadataConfig         `yaml:"metadata"`
	Request        RequestConfig          `yaml:"request"`
	Assertions     []AssertionConfig      `yaml:"assertions"`
	TimeoutSeconds int                    `yaml:"timeout_seconds"`
}

// MetadataConfig holds optional metadata for the test
type MetadataConfig struct {
	Name string `yaml:"name"`
	Tags []string `yaml:"tags"`
}

// RequestConfig represents the HTTP request configuration
type RequestConfig struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

// AssertionConfig represents a single assertion
type AssertionConfig struct {
	Type     string      `yaml:"type"`
	Path     string      `yaml:"path"`
	Expected interface{} `yaml:"expected"`
}

// Result holds the result of a test execution
type Result struct {
	Name         string
	Pass         bool
	StatusCode   int
	ResponseTime time.Duration
	ResponseBody string
	Assertions   []AssertionResult
	Error        error
}

// AssertionResult holds the result of a single assertion
type AssertionResult struct {
	Type     string
	Expected interface{}
	Actual   interface{}
	Pass     bool
	Message  string
}

// Execute performs the HTTP request and validates assertions
func Execute(config TestConfig) Result {
	result := Result{Name: config.Metadata.Name}

	// Default timeout
	timeout := 10 * time.Second
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	client := &http.Client{Timeout: timeout}

	// Build request
	method := strings.ToUpper(config.Request.Method)
	if method == "" {
		method = "GET"
	}

	var body *strings.Reader
	if config.Request.Body != "" {
		body = strings.NewReader(config.Request.Body)
	} else {
		body = strings.NewReader("")
	}

	req, err := http.NewRequest(method, config.Request.URL, body)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// Set headers
	for key, value := range config.Request.Headers {
		req.Header.Set(key, value)
	}

	// Execute request with timing
	start := time.Now()
	resp, err := client.Do(req)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		return result
	}
	responseBody := string(bodyBytes)
	result.ResponseBody = responseBody

	// Run assertions
	result.Assertions = make([]AssertionResult, 0, len(config.Assertions))
	allPassed := true

	for _, assertion := range config.Assertions {
		assertionResult := validateAssertion(assertion, result.StatusCode, responseBody)
		result.Assertions = append(result.Assertions, assertionResult)
		if !assertionResult.Pass {
			allPassed = false
		}
	}

	result.Pass = allPassed
	return result
}