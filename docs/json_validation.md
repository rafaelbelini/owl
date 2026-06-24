# JSON Validation

The `json_path` assertion allows you to validate specific fields within a JSON response body.

## YAML Schema

```yaml
assertions:
  - type: "json_path"
    path: "<jsonpath-expression>"
    expected: "<expected-value>"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Must be `"json_path"` |
| `path` | string | Yes | JSONPath expression to extract the value |
| `expected` | any | Yes | Expected value at the specified path |

## JSONPath Expressions

Standard JSONPath syntax is supported:

| Expression | Description |
|------------|-------------|
| `$.field` | Access root object field |
| `$.nested.field` | Access nested fields |
| `$.array[0]` | Access first element of array |
| `$.array[-1]` | Access last element of array |
| `$.field[]` | All elements of array under field |

## Examples

### Basic Field Validation

```yaml
name: "Validate User Response"
request:
  url: "https://api.example.com/users/1"
  method: GET

assertions:
  - type: "status_code"
    expected: 200
  - type: "json_path"
    path: "$.name"
    expected: "John Doe"
```

### Array Element Validation

```yaml
assertions:
  - type: "json_path"
    path: "$.slideshow.slides[0].title"
    expected: "Wake up to WonderWidgets!"
```

### Nested Object Validation

```yaml
assertions:
  - type: "json_path"
    path: "$.data.user.profile.email"
    expected: "user@example.com"
```

## Example Test File

```yaml
version: 1
metadata:
  name: "JSON Path Validation Test"
  tags:
    - jsonpath
request:
  url: "https://httpbin.org/json"
  method: GET
  headers:
    Accept: "application/json"
  body: ""

assertions:
  - type: "status_code"
    expected: 200
  - type: "json_path"
    path: "$.slideshow.author"
    expected: "Yours Truly"
  - type: "json_path"
    path: "$.slideshow.slides[0].title"
    expected: "Wake up to WonderWidgets!"

timeout_seconds: 10
```

## Output

**Pass:**
```
✓ json_path: path "$.slideshow.author" == "Yours Truly"
```

**Fail:**
```
✗ json_path: path "$.slideshow.title": expected "Wrong Title", got "Sample Slide Show"
```

## Error Cases

| Error | Cause |
|-------|-------|
| `invalid JSON response` | Response body is not valid JSON |
| `path "..." not found in JSON` | JSONPath expression doesn't match any value |
