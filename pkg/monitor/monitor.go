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
	Name           string                 `yaml:"name"`
	Request        RequestConfig          `yaml:"request"`
	Assertions     []AssertionConfig      `yaml:"assertions"`
	TimeoutSeconds int                    `yaml:"timeout_seconds"`
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
	Expected interface{} `yaml:"expected"`
}

// Result holds the result of a test execution
type Result struct {
	Name         string
	Pass         bool
	StatusCode   int
	ResponseTime time.Duration
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
	result := Result{Name: config.Name}

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

// validateAssertion checks a single assertion against the response
func validateAssertion(assertion AssertionConfig, statusCode int, body string) AssertionResult {
	result := AssertionResult{
		Type:     assertion.Type,
		Expected: assertion.Expected,
	}

	switch assertion.Type {
	case "status_code":
		expectedCode, ok := assertion.Expected.(int)
		if !ok {
			// Try float64 (YAML parses ints as float64 in some cases)
			if f, ok := assertion.Expected.(float64); ok {
				expectedCode = int(f)
			}
		}
		result.Actual = statusCode
		result.Pass = statusCode == expectedCode
		result.Message = fmt.Sprintf("status code: expected %d, got %d", expectedCode, statusCode)

	case "contains_text":
		expectedText, ok := assertion.Expected.(string)
		if !ok {
			result.Pass = false
			result.Message = "contains_text expected value must be a string"
			return result
		}
		result.Actual = "\"" + expectedText + "\" in body"
		result.Pass = strings.Contains(body, expectedText)
		if result.Pass {
			result.Message = fmt.Sprintf("found text %q in body", expectedText)
		} else {
			result.Message = fmt.Sprintf("text %q not found in body", expectedText)
		}

	default:
		result.Pass = false
		result.Message = fmt.Sprintf("unknown assertion type: %s", assertion.Type)
	}

	return result
}