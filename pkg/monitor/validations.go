package monitor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveagle/jsonpath"
)

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

	case "json_path":
		path := assertion.Path
		if path == "" {
			result.Pass = false
			result.Message = "json_path requires a 'path' field"
			return result
		}

		// Parse JSON body
		var data interface{}
		if err := json.Unmarshal([]byte(body), &data); err != nil {
			result.Pass = false
			result.Message = fmt.Sprintf("invalid JSON response: %v", err)
			return result
		}

		// Evaluate JSON path
		actualValue, err := jsonpath.JsonPathLookup(data, path)
		if err != nil {
			result.Pass = false
			result.Message = fmt.Sprintf("path %q not found in JSON", path)
			return result
		}

		result.Actual = actualValue

		// Compare with expected value
		expectedStr := fmt.Sprintf("%v", assertion.Expected)
		actualStr := fmt.Sprintf("%v", actualValue)
		result.Pass = actualStr == expectedStr

		if result.Pass {
			result.Message = fmt.Sprintf("path %q == %q", path, expectedStr)
		} else {
			result.Message = fmt.Sprintf("path %q: expected %q, got %q", path, expectedStr, actualStr)
		}

	default:
		result.Pass = false
		result.Message = fmt.Sprintf("unknown assertion type: %s", assertion.Type)
	}

	return result
}
