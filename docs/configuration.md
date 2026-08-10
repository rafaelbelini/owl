# Configuration Files

Owl CLI supports configuration files to define reusable values that can be referenced in test files using placeholders.

## Overview

Configuration files allow you to:
- Define values once and reuse them across multiple tests
- Separate sensitive data (like API tokens) from test definitions
- Change configuration per environment without modifying test files
- Scope configurations to specific directories
- Define global pre/post scripts that run before or after all tests

## File Format

Configuration files are named `owl.config` and use a simple `key=value` format:

```bash
# Comment
api_base_url=https://api.example.com
api_token=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
test_email=test@example.com

# Global scripts (optional)
before_all_script=./scripts/setup.sh
after_all_script=./scripts/teardown.sh
```

### Rules

- Each line must be `key=value` format
- Lines starting with `#` are comments and ignored
- Empty lines are ignored
- Keys cannot be empty
- Values are strings (no quotes needed)

## Using Placeholders

In your test YAML files, reference configuration values using `${key_name}` syntax:

```yaml
request:
  url: "${api_base_url}/users/profile"
  headers:
    Authorization: "${api_token}"
```

### Example: HTTP Test with Config

**owl.config:**
```bash
api_base_url=https://api.example.com
api_token=Bearer my_secret_token
```

**test.yaml:**
```yaml
version: 1

metadata:
  name: "API Test"

request:
  url: "${api_base_url}/users"
  method: GET
  headers:
    Authorization: "${api_token}"

assertions:
  - type: "status_code"
    expected: 200
```

### Example: Browser Test with Config

**owl.config:**
```bash
base_url=http://localhost:8000
test_email=test@example.com
test_password=secret123
```

**test.yaml:**
```yaml
version: 1

metadata:
  name: "Login Test"

browser:
  url: "${base_url}/login.html"
  headless: true

steps:
  - name: "Fill login form"
    action: "fill"
    selector: "input[name='email']"
    value: "${test_email}"

  - name: "Fill password"
    action: "fill"
    selector: "input[name='password']"
    value: "${test_password}"
```

## Directory Scoping

Configuration files follow a scoping hierarchy based on directory structure:

### Basic Scoping

When running tests from a directory, the `owl.config` file in that directory applies to all tests in it and its subdirectories.

```
tests/
├── owl.config   # Applied to all tests
├── test1.yaml
└── test2.yaml
```

### Nested Scoping

For nested directories, each directory can have its own `owl.config`. More specific (deeper) configurations override less specific ones.

```
tests/
├── owl.config   # Applied to all tests (base config)
├── test1.yaml
└── subdir/
    ├── owl.config    # Applied to tests in subdir/ (overrides parent)
    └── test2.yaml    # Uses subdir/owl.config
```

**Example:**

`tests/owl.config`:
```bash
api_url=https://api.example.com
api_token=Bearer parent_token
environment=production
```

`tests/subdir/owl.config`:
```bash
api_token=Bearer child_token  # Override parent value
environment=staging          # Override parent value
```

**test1.yaml** (in `tests/`):
- `api_url`: `https://api.example.com` (from parent)
- `api_token`: `Bearer parent_token` (from parent)
- `environment`: `production` (from parent)

**test2.yaml** (in `tests/subdir/`):
- `api_url`: `https://api.example.com` (inherited from parent)
- `api_token`: `Bearer child_token` (overridden in child)
- `environment`: `staging` (overridden in child)

## Global Scripts

You can define global scripts that run before and after all tests in the `owl.config` file:

### before_all_script

A script that runs once before any tests in the scope are executed. Use this for setup tasks like:
- Starting services
- Initializing databases
- Creating test fixtures

### after_all_script

A script that runs once after all tests in the scope have completed (even if some tests failed). Use this for cleanup tasks like:
- Stopping services
- Cleaning up temporary files
- Generating reports

**Example:**

```bash
# owl.config
before_all_script=./scripts/setup.sh
after_all_script=./scripts/teardown.sh
```

**scripts/setup.sh:**
```bash
#!/bin/bash
echo "Setting up test environment..."
docker-compose up -d
```

**scripts/teardown.sh:**
```bash
#!/bin/bash
echo "Cleaning up test environment..."
docker-compose down
```

### Script Execution

- Scripts are executed in the directory containing the `owl.config` file
- If a `before_all_script` fails, no tests are run and the test suite fails
- If an `after_all_script` fails, a warning is shown but the test results are not affected
- Scripts support a 30-second default timeout (configurable via `script_timeout` in test YAML)

## Error Handling

### Missing Keys

If a test references a placeholder that doesn't exist in any configuration file, an error is thrown:

```
Error: placeholder '${missing_key}' not found in configuration
```

**Important:** In batch execution (running all tests in a directory), a missing key error will cause that specific test to fail, but execution will continue for other tests.

### Missing Config Files

If no `owl.config` file exists in the test file's directory or any parent directory, the test runs normally with no placeholder resolution.

## Common Use Cases

### 1. API Authentication

```bash
# owl.config
api_token=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

```yaml
# test.yaml
request:
  headers:
    Authorization: "${api_token}"
```

### 2. Multiple Environments

```bash
# production owl.config
api_url=https://api.production.com
db_host=prod-db.example.com

# staging owl.config
api_url=https://api.staging.com
db_host=staging-db.example.com
```

### 3. Test Data

```bash
# owl.config
test_user_email=user@test.com
test_user_id=12345
test_product_id=SKU-001
```

## Best Practices

1. **Use `.gitignore`**: Add `owl.config` to `.gitignore` if it contains sensitive data

2. **Create a template**: Use `owl.config.example` or `owl.config.template` as a template showing required keys without actual secrets

3. **Organize by environment**: Use directory structure to separate environment configs:
   ```
   tests/
   ├── production/
   │   └── owl.config
   ├── staging/
   │   └── owl.config
   └── local/
       └── owl.config
   ```

4. **Document your keys**: Add comments in config files explaining what each key is used for

## Running Tests with Config

Configuration files are loaded automatically. No additional flags are needed:

```bash
# Single test - loads config from test file's directory
owl run tests/api_test.yaml

# All tests in directory - loads configs from all directories
owl run tests/

# With other flags
owl run tests/ --silent --report=json
```
