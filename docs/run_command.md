# Run Command

Executes API monitoring tests defined in YAML files.

## Usage

```bash
owl run <path>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `<path>` | Path to a `.yaml`/`.yml` file or directory containing such files |

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--verbose` | `-v` | `false` | Show response body on failure for debugging |
| `--silent` | `-s` | `false` | Suppress individual test output, show only summary |
| `--report` | - | `""` | Generate report in specified format (html, json) |

## Examples

### Single file

```bash
owl run tests/check-api.yaml
```

### Directory (recursive)

```bash
owl run tests/
```

### Silent mode

```bash
owl run tests/ --silent
```

### Generate HTML report

```bash
owl run tests/ --report=html
```

### Generate JSON report

```bash
owl run tests/ --report=json
```

## Exit Codes

| Code | Description |
|------|-------------|
| `0`  | All tests passed |
| `1`  | One or more tests failed |

## Output Format

```
Found 5 test file(s)

Test Name
==========
  ✓ PASS [200] 245ms
    ✓ status_code: status code: expected 200, got 200
    ✓ contains_text: found text "success" in body

SUMMARY
----------------------------------------
Total:  5
Passed: 3
Failed: 2
```

### Status Indicators

- **✓ PASS** — Test passed (green)
- **✗ FAIL** — Test failed assertion(s) (red)
- **✗ ERROR** — Test encountered an error (red)

## Reports

When using the `--report` flag, a report is generated in the `.results/` directory.

### Report Formats

| Format | File Extension | Description |
|--------|----------------|-------------|
| `html` | `.html` | Interactive HTML report with dashboard and filters |
| `json` | `.json` | JSON report for integration with other tools |

### Report Location

Reports are saved to `.results/` (hidden directory, git-ignored):

```
.results/
├── http_2026-08-05_14-30-00.html
├── http_2026-08-05_14-30-00.json
├── browser_2026-08-05_14-30-15.html
└── browser_2026-08-05_14-30-15.json
```

### HTML Report Features

- **Dashboard metrics**: Total, passed, failed, skipped counts
- **Interactive filters**: Search by test name/ID, filter by status
- **Detailed results table**: Test ID, suite, name, duration, status
- **Error logs**: Failed tests show error details

### JSON Report Structure

```json
{
  "executed_at": "2026-08-05T14:30:00Z",
  "type": "http",
  "summary": {
    "total": 5,
    "passed": 3,
    "failed": 2,
    "skipped": 0
  },
  "tests": [
    {
      "id": "HTTP-1",
      "suite": "HTTP Tests",
      "name": "User Profile API",
      "duration": "245ms",
      "status": "pass"
    }
  ]
}
```

### Report Types

| Type | Description |
|------|-------------|
| HTTP | HTTP API test results (filename: `http_YYYY-MM-DD_HH-MM-SS.ext`) |
| Browser | Browser automation results (filename: `browser_YYYY-MM-DD_HH-MM-SS.ext`) |
| Mixed | When running a directory with both test types, separate reports are generated for each |

## Finding Test Files

When given a directory, the CLI recursively walks the directory tree and collects all files matching:

- `*.yaml`
- `*.yml`

## Timeout

Each test can define its own timeout via `timeout_seconds` in the YAML. Default is 10 seconds if not specified.

## Retry Configuration

Tests can be configured to automatically retry on transient errors such as timeouts or network issues.

### HTTP Tests

```yaml
version: 1

metadata:
  name: "API Test with Retry"

retries: 3           # Number of retries on failure (default: 0)
retry_interval: 1000 # Interval between retries in milliseconds (default: 0)

request:
  url: "https://api.example.com/users"
  method: GET

assertions:
  - type: "status_code"
    expected: 200
```

### Browser Tests

```yaml
version: 1

metadata:
  name: "Browser Test with Retry"

retries: 2           # Number of retries on failure (default: 0)
retry_interval: 500  # Interval between retries in milliseconds (default: 0)

browser:
  url: "https://example.com"
  headless: true

steps:
  - name: "Navigate"
    action: "goto"
    url: "https://example.com"

assertions:
  - type: "url_contains"
    path: "example.com"
```

### Retry Behavior

| Condition | Retry? | Reason |
|-----------|--------|--------|
| Timeout error | Yes | Transient network issue |
| Connection refused | Yes | Service temporarily unavailable |
| Assertion failure | No | Test logic failure, retry won't help |
| HTTP 4xx/5xx response | No | Server returned an error response |
| Invalid selector | No | Element not found, retry won't help |

### Retryable Errors

The following errors trigger a retry:

- **Timeout errors**: `context deadline exceeded`, `i/o timeout`
- **Network errors**: `connection refused`, `connection reset`, `no such host`, `network is unreachable`
- **Browser errors**: Navigation timeout, page load timeout

### Default Values

```yaml
retries: 0         # No retries by default
retry_interval: 0  # No delay between retries
```
