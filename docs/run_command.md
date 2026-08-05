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

## Examples

### Single file

```bash
owl run tests/check-api.yaml
```

### Directory (recursive)

```bash
owl run tests/
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

## Finding Test Files

When given a directory, the CLI recursively walks the directory tree and collects all files matching:

- `*.yaml`
- `*.yml`

## Timeout

Each test can define its own timeout via `timeout_seconds` in the YAML. Default is 10 seconds if not specified.
