# Configuration Files

Owl CLI supports configuration files to definir valores reutilizáveis que podem ser referenciados em arquivos de teste usando placeholders.

## Overview

Configuration files allow you to:
- Define values once and reuse them across multiple tests
- Separate sensitive data (like API tokens) from test definitions
- Change configuration per environment without modifying test files
- Scope configurations to specific directories
- Define global pre/post scripts that run before or after all tests
- Use separate env files for different environments (production, staging, etc.)

## File Structure

Owl uses two types of configuration files:

| File | Purpose | Contains |
|------|---------|----------|
| `owl.config` | Test execution configuration | `before_all_script`, `after_all_script`, `env` |
| `.env` (or custom) | Test variables | API URLs, tokens, credentials, etc. |

### owl.config

Contains only execution configuration:

```bash
# Scripts (optional)
before_all_script=./scripts/setup.sh
after_all_script=./scripts/teardown.sh

# Env file path (optional)
# If not specified, .env will be auto-discovered in the test directory if it exists
env=.env
```

### Env File

Contains test variables that are used as placeholders:

```bash
# API Configuration
api_base_url=https://api.example.com
api_token=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
test_email=test@example.com
```

### Auto-discovery of .env

If no `env` key is configured in `owl.config`, Owl will automatically look for a `.env` file in the test directory:

```
tests/
├── owl.config    # No env key - .env will be auto-discovered
├── .env          # Automatically loaded if exists
└── test.yaml
```

If the `.env` file doesn't exist, no variables are loaded and tests run normally without placeholder resolution.

## Using Placeholders

In your test YAML files, reference configuration values using `${key_name}` syntax:

```yaml
request:
  url: "${api_base_url}/users/profile"
  headers:
    Authorization: "${api_token}"
```

### Example: HTTP Test with Config

**.env:**
```bash
api_base_url=https://api.example.com
api_token=Bearer my_secret_token
```

**owl.config:**
```bash
env=.env
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

**.env:**
```bash
base_url=http://localhost:8000
test_email=test@example.com
test_password=secret123
```

**owl.config:**
```bash
env=.env
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

## Multiple Environments

One of the main benefits of this structure is easy environment switching.

### Structure

```
tests/
├── owl.config        # env=.env (or .env.production)
├── .env             # Development variables
├── .env.production  # Production variables
└── test.yaml
```

### Changing Environments

Simply change the `env` value in `owl.config`:

```bash
# Development (auto-discovers .env)
# env=.env  <-- can be omitted if using .env

# Explicitly specify .env
env=.env

# Production
env=.env.production

# Staging
env=.env.staging
```

**Note:** If `env` is explicitly configured in `owl.config`, only that file is loaded. If a different `.env` file also exists in the directory, it is ignored.

Each env file can have different values:

**.env:**
```bash
api_base_url=http://localhost:3000
```

**.env.production:**
```bash
api_base_url=https://api.production.com
```

## Directory Scoping

Configuration files follow a scoping hierarchy based on directory structure.

### Basic Scoping

When running tests from a directory, the `owl.config` and `.env` files in that directory apply to all tests in it and its subdirectories.

```
tests/
├── owl.config   # Applied to all tests
├── .env         # Variables for all tests
├── test1.yaml
└── test2.yaml
```

### Nested Scoping

For nested directories, each directory can have its own `owl.config` and `.env` file. Scripts (`before_all_script`, `after_all_script`) are scoped per directory. Env files are **not** merged from parent to child directories.

```
tests/
├── owl.config   # Scripts for tests/ and subdir/
├── .env         # Only for tests/ directory
├── test1.yaml
└── subdir/
    ├── owl.config   # Scripts for subdir/ only
    ├── .env         # Only for subdir/ directory
    └── test2.yaml
```

**Example:**

`tests/owl.config`:
```bash
before_all_script=./scripts/setup.sh
```

`tests/.env`:
```bash
api_url=https://api.example.com
api_token=Bearer token123
```

`tests/subdir/owl.config`:
```bash
# No env key - .env will be auto-discovered in subdir/
before_all_script=./other_setup.sh
```

`tests/subdir/.env`:
```bash
api_url=https://staging.example.com
```

**test1.yaml** (in `tests/`):
- Uses `tests/.env` for variables

**test2.yaml** (in `tests/subdir/`):
- Uses `tests/subdir/.env` for variables (auto-discovered)
- api_url from subdir/.env, not parent

## Global Scripts

You can define global scripts that run before and after all tests in the `owl.config` file.

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

## Rules

### owl.config

- Each line must be `key=value` format
- Lines starting with `#` are comments and ignored
- Empty lines are ignored
- Keys cannot be empty
- Supported keys: `before_all_script`, `after_all_script`, `env`
- All other keys are ignored (use the env file instead)

### Env File

- Each line must be `key=value` format
- Lines starting with `#` are comments and ignored
- Empty lines are ignored
- Keys cannot be empty
- Values are strings (no quotes needed)

## Error Handling

### Missing Keys

If a test references a placeholder that doesn't exist in any env file, an error is thrown:

```
Error: placeholder '${missing_key}' not found in configuration
```

**Important:** In batch execution (running all tests in a directory), a missing key error will cause that specific test to fail, but execution will continue for other tests.

### Missing Config Files

If no `owl.config` file exists in the test file's directory or any parent directory, the test runs normally with no placeholder resolution.

If the `env` key is not specified, no env file is loaded and only hardcoded values work.

### Missing Env File

If the env file specified in `owl.config` doesn't exist, the test will fail with an error:

```
Error: failed to load env file /path/to/.env: open /path/to/.env: no such file or directory
```

## Common Use Cases

### 1. API Authentication

**.env:**
```bash
api_token=Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**test.yaml:**
```yaml
request:
  headers:
    Authorization: "${api_token}"
```

### 2. Multiple Environments

**.env:**
```bash
api_url=http://localhost:3000
db_host=localhost
```

**.env.production:**
```bash
api_url=https://api.production.com
db_host=prod-db.example.com
```

**.env.staging:**
```bash
api_url=https://api.staging.com
db_host=staging-db.example.com
```

### 3. Test Data

**.env:**
```bash
test_user_email=user@test.com
test_user_id=12345
test_product_id=SKU-001
```

## Best Practices

1. **Use `.gitignore`**: Add `.env` to `.gitignore` if it contains sensitive data. Keep `owl.config` tracked as it doesn't contain secrets.

2. **Create templates**: Use `.env.example` or `.env.template` as a template showing required keys without actual secrets:
   ```bash
   cp .env .env.example
   ```

3. **Organize by environment**: Use directory structure to separate environment configs:
   ```
   tests/
   ├── owl.config           # env=.env
   ├── .env                 # Development
   ├── production/
   │   ├── owl.config       # env=../.env.production
   │   └── .env.production
   └── staging/
       ├── owl.config       # env=../.env.staging
       └── .env.staging
   ```

4. **Document your keys**: Add comments in env files explaining what each key is used for:
   ```bash
   # API Authentication Token
   api_token=Bearer your_token_here
   ```

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
