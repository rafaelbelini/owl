# Owl CLI

A high-performance, concurrent API monitoring CLI application built with Go and Cobra.

## Features

- **Single & Batch Testing**: Run tests from a single YAML file or an entire directory
- **HTTP Methods**: Support for GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **Custom Headers**: Define custom headers including authentication tokens
- **Response Assertions**:
  - `status_code`: Validate HTTP status codes
  - `contains_text`: Check for text presence in response body
  - `json_path`: Validate JSON values using JSONPath expressions, including array indexing and filter expressions
- **Colored Output**: Green for passes, red for failures
- **Verbose Mode**: Show response body on failure for easier debugging (`-v`)
- **Response Timing**: Track request duration for performance monitoring
- **Error Handling**: Robust handling of network timeouts, malformed YAML, and invalid paths

## Installation

```bash
# Clone the repository
git clone https://github.com/rafael/owl.git
cd owl/cli

# Build the CLI
go build -o owl .

# Optional: Add to PATH
mv owl /usr/local/bin/
```

## Quick Start

### Create a test file

Create a YAML file (e.g., `check-api.yaml`):

```yaml
version: 1
metadata:
  name: "User Profile API Check"
  tags:
    - user
    - api
request:
  url: "https://api.example.com/v1/user/profile"
  method: GET
  headers:
    Authorization: "Bearer your-test-token"
    Accept: "application/json"
  body: ""

assertions:
  - type: "status_code"
    expected: 200
  - type: "contains_text"
    expected: "username"

timeout_seconds: 10
```

### Run tests

**Single file:**
```bash
./owl run check-api.yaml
```

**Directory:**
```bash
./owl run ./tests/
```

## YAML Schema Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | int | Yes | Schema version (currently `1`) |
| `metadata.name` | string | Yes | Name of the test |
| `metadata.tags` | array | No | Tags for categorizing tests |
| `request.url` | string | Yes | Full URL to call |
| `request.method` | string | No | HTTP method (default: GET) |
| `request.headers` | map | No | Custom headers |
| `request.body` | string | No | Request body |
| `assertions` | array | Yes | List of assertions |
| `timeout_seconds` | int | No | Request timeout (default: 10) |

### Assertion Types

#### status_code
```yaml
- type: "status_code"
  expected: 200
```

#### contains_text
```yaml
- type: "contains_text"
  expected: "success"
```

#### json_path
```yaml
- type: "json_path"
  path: "$.user.name"
  expected: "John"
```

**Array indexing:**
```yaml
- type: "json_path"
  path: "$.users[0].name"
  expected: "Alice"
- type: "json_path"
  path: "$.items[-1].id"
  expected: 42
```

## Example Output

```
Found 2 test file(s)

User Profile API Check
========================
  ✓ PASS [200] 245ms
    ✓ status_code: status code: expected 200, got 200
    ✓ contains_text: found text "username" in body

================================
SUMMARY
----------------------------------------
Total:  2
Passed: 2
Failed: 0
```

## License

MIT