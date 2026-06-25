package monitor

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oliveagle/jsonpath"
)

// filterRegex matches {"key":"value"} or {"key":value} patterns
var filterRegex = regexp.MustCompile(`\{\s*"([^"]+)"\s*:\s*"([^"]*)"\s*\}`)

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

		// Check if path contains filter syntax
		if filterRegex.MatchString(path) {
			actualValue, err := evaluateFilteredPath(data, path)
			if err != nil {
				result.Pass = false
				result.Message = err.Error()
				return result
			}
			result.Actual = actualValue
			expectedStr := fmt.Sprintf("%v", assertion.Expected)
			actualStr := fmt.Sprintf("%v", actualValue)
			result.Pass = actualStr == expectedStr
			if result.Pass {
				result.Message = fmt.Sprintf("path %q == %q", path, expectedStr)
			} else {
				result.Message = fmt.Sprintf("path %q: expected %q, got %q", path, expectedStr, actualStr)
			}
		} else {
			// Standard JSONPath
			actualValue, err := jsonpath.JsonPathLookup(data, path)
			if err != nil {
				result.Pass = false
				result.Message = fmt.Sprintf("path %q not found in JSON", path)
				return result
			}
			result.Actual = actualValue
			expectedStr := fmt.Sprintf("%v", assertion.Expected)
			actualStr := fmt.Sprintf("%v", actualValue)
			result.Pass = actualStr == expectedStr
			if result.Pass {
				result.Message = fmt.Sprintf("path %q == %q", path, expectedStr)
			} else {
				result.Message = fmt.Sprintf("path %q: expected %q, got %q", path, expectedStr, actualStr)
			}
		}

	default:
		result.Pass = false
		result.Message = fmt.Sprintf("unknown assertion type: %s", assertion.Type)
	}

	return result
}

// evaluateFilteredPath handles JSONPath with filter expressions like {"key":"value"}
func evaluateFilteredPath(data interface{}, path string) (interface{}, error) {
	// Remove leading $ if present
	path = strings.TrimPrefix(path, "$.")

	// Find all filter patterns and split the path
	parts := filterRegex.Split(path, -1)

	// Find all matches to get filter key-value pairs
	matches := filterRegex.FindAllStringSubmatch(path, -1)

	current := data

	// Navigate through the path parts
	for i, part := range parts {
		// Navigate through regular path segments
		segments := strings.Split(strings.Trim(part, "."), ".")
		for _, seg := range segments {
			if seg == "" {
				continue
			}
			current = navigate(current, seg)
			if current == nil {
				return nil, fmt.Errorf("path segment %q not found", seg)
			}
		}

		// Apply filter if there's a match for this position
		if i < len(matches) {
			key := matches[i][1]
			value := matches[i][2]

			// Check if current is an array - filter within array
			if arr, ok := current.([]interface{}); ok {
				var matchedItem interface{}
				matchCount := 0

				for _, item := range arr {
					itemMap, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					if itemKey, exists := itemMap[key]; exists {
						if fmt.Sprintf("%v", itemKey) == value {
							matchedItem = item
							matchCount++
							if matchCount > 1 {
								return nil, fmt.Errorf("multiple matches (%d) found for filter {%q:%q}, expected exactly 1", matchCount, key, value)
							}
						}
					}
				}
				if matchCount == 0 {
					return nil, fmt.Errorf("no item found matching filter {%q:%q}", key, value)
				}
				current = matchedItem
			} else if m, ok := current.(map[string]interface{}); ok {
				// Check if the map itself matches the filter
				if itemKey, exists := m[key]; exists {
					if fmt.Sprintf("%v", itemKey) != value {
						return nil, fmt.Errorf("map filter {%q:%q} did not match", key, value)
					}
					// Map matches filter, continue with same map (no change to current)
				} else {
					return nil, fmt.Errorf("key %q not found in map for filter", key)
				}
			} else {
				return nil, fmt.Errorf("filter cannot be applied to %T", current)
			}
		}
	}

	return current, nil
}

// navigate moves through data based on a single segment (field name or array index)
func navigate(data interface{}, segment string) interface{} {
	if data == nil {
		return nil
	}

	// Check if it's an array index
	if segment[0] >= '0' && segment[0] <= '9' {
		arr, ok := data.([]interface{})
		if !ok {
			return nil
		}
		idx := 0
		fmt.Sscanf(segment, "%d", &idx)
		if idx < 0 || idx >= len(arr) {
			return nil
		}
		return arr[idx]
	}

	// Otherwise it's a field name
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	return m[segment]
}
