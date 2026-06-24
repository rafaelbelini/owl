# Coding Guidelines

## Project Structure

```
owl/cli/
├── main.go              # Entry point - minimal, only imports and calls cmd.Execute()
├── cmd/                 # Cobra commands
│   ├── root.go          # Root command definition
│   └── run.go           # run subcommand
├── pkg/
│   └── monitor/         # Core monitoring logic
│       ├── monitor.go   # HTTP execution & assertions
│       ├── loader.go    # YAML loading & file discovery
│       └── output.go    # Console output formatting
└── docs/                # Documentation
```

## Go Conventions

### Package Naming

- Use short, meaningful names: `monitor`, not `apimonitor`
- Avoid `util` or `common` packages
- Group related functionality in the same package

### Error Handling

- Return errors rather than logging and continuing (except at top-level)
- Use `fmt.Errorf("action: %w", err)` for wrapped errors
- Never silently ignore errors with `_`

```go
// Good
data, err := os.ReadFile(path)
if err != nil {
    return nil, fmt.Errorf("failed to read file: %w", err)
}

// Bad
data, _ := os.ReadFile(path)
```

### Naming

- **Variables/Functions**: `camelCase`
- **Types/Structs**: `PascalCase`
- **Constants**: `PascalCase` or `ALL_CAPS` for exported constants
- **Package names**: lowercase, no underscores

### Imports

Group imports with `go fmt`:

1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/fatih/color"
    "github.com/rafael/owl/cli/pkg/monitor"
)
```

## Cobra Commands

### Command Structure

```go
var runCmd = &cobra.Command{
    Use:   "run",
    Short: "One-liner description",
    Long:  `Multi-line description if needed.`,
    Args:  cobra.ExactArgs(1),  // Validate arguments
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil  // or error
    },
}
```

### Argument Validation

Use appropriate `Args` validators:

| Validator | Use Case |
|-----------|----------|
| `cobra.ExactArgs(n)` | Exactly n arguments required |
| `cobra.MinArgs(n)` | At least n arguments |
| `cobra.NoArgs` | No arguments allowed |
| `cobra.OnlyValidArgs` | Args must be in ValidArgs list |

## YAML Configuration

### Schema Design

```yaml
name: "Test Name"                    # Required: string
request:
  url: "https://..."                 # Required: string
  method: GET                        # Optional: defaults to GET
  headers: {}                        # Optional: map[string]string
  body: ""                           # Optional: string
assertions: []                       # Required: array
  - type: "status_code"              # Required: assertion type
    expected: <value>                # Required: expected value
timeout_seconds: 10                  # Optional: default 10
```

### Field Types

| YAML Type | Go Type | Notes |
|-----------|---------|-------|
| string | `string` | Direct mapping |
| number | `int` or `float64` | YAML parses ints as float64 in some cases |
| array | `[]interface{}` | Unmarshal and type-assert |
| map | `map[string]interface{}` | Direct mapping |

## Output & Formatting

### Colored Output

Use `fatih/color` for colored console output:

```go
import "github.com/fatih/color"

green := color.New(color.FgGreen)
red := color.New(color.FgRed)
bold := color.New(color.Bold)

green.Printf("  ✓ PASS\n")
red.Printf("  ✗ FAIL: %v\n", err)
```

### Print Result Structure

```
Test Name
==========
  ✓ PASS [200] 245ms
    ✓ status_code: status code: expected 200, got 200
```

### Summary Format

```
SUMMARY
----------------------------------------
Total:  5
Passed: 3
Failed: 2
```

## Testing

### Test Files Location

- Test YAML files go in `tests/` directory
- Use descriptive names: `check-api.yaml`, `health-check.yaml`
- Include both passing and failing test examples

### Manual Testing

```bash
# Single test
./owl run tests/check-api.yaml

# All tests
./owl run tests/

# Verify exit code
./owl run tests/; echo "Exit code: $?"
```

## Documentation

### README.md

- Features list
- Installation instructions
- Quick start guide
- YAML schema reference
- Examples

### Doc Files (docs/)

- `run_command.md` — CLI command usage
- `json_validation.md` — Feature-specific documentation
- `CODING_GUIDELINES.md` — This file

### Commit Messages

Follow Conventional Commits:

```
feat: add json_path assertion type
fix: correct failed test count in summary
docs: update README with new examples
```

## Dependencies

### Adding Dependencies

```bash
go get <package>@latest
```

### Version Management

- Use `@latest` to get latest stable version
- Commit `go.mod` and `go.sum` together

## Build & Release

### Build Binary

```bash
go build -o owl .
```

### Cross-Compile

```bash
GOOS=linux GOARCH=amd64 go build -o owl-linux .
GOOS=darwin GOARCH=amd64 go build -o owl-macos .
GOOS=windows GOARCH=amd64 go build -o owl.exe .
```
