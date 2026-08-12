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
  expected: o00
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

## Browser Tests

Owl CLI também suporta testes de browser usando [go-rod](https://github.com/go-rod/rod) para automação de Chrome/Edge. Permite testar fluxos de utilizador completos, incluindo interação com elementos, submissão de formulários, e verificação de estado.

### Pré-requisitos

- Chrome ou Chromium instalado
- Servidor web local para servir ficheiros HTML (ex: `npx serve examples/browser -p 8000`)

### Quick Start

```bash
# 1. Iniciar servidor local
npx serve examples/browser -p 8000

# 2. Executar testes
go run . run examples/browser/

# 3. Com verbose (mostrar detalhes)
go run . run examples/browser/login_test.yaml --verbose

# 4. Silent mode (apenas summary)
go run . run examples/browser/ --silent
```

### Exemplo: Login Flow

```yaml
version: 1

metadata:
  name: "Login Flow Test"
  description: "Tests the login flow"
  tags:
    - login
    - smoke

browser:
  url: "http://localhost:8000/login.html"
  headless: true
  timeout_seconds: 30
  wait_until: "networkidle2"
  viewport:
    width: 1280
    height: 720

steps:
  - name: "Fill email field"
    action: "fill"
    selector: "input[name='email']"
    value: "test@example.com"

  - name: "Fill password field"
    action: "fill"
    selector: "input[name='password']"
    value: "KFwBNLfFXn71rCb5QJuU"

  - name: "Submit form"
    action: "click"
    selector: "#submit-btn"

  - name: "Wait for redirect"
    action: "wait_url"
    pattern: "dashboard"

assertions:
  - name: "URL contains dashboard"
    type: "url_contains"
    path: "dashboard"

  - name: "Title is correct"
    type: "title"
    expected: "Dashboard | Example App"

  - name: "Token saved"
    type: "local_storage"
    key: "auth_token"
```

### Actions (Steps)

| Action | Parâmetros | Descrição |
|--------|------------|-----------|
| `goto` | `url` | Navega para URL |
| `click` | `selector` | Clica em elemento |
| `fill` / `fill_text` | `selector`, `value` | Preenche campo de texto |
| `hover` | `selector` | Move mouse sobre elemento |
| `select` | `selector`, `value` | Seleciona opção em dropdown |
| `press` | `selector`, `key` | Pressiona tecla (Enter, Escape, Tab, etc.) |
| `wait_url` | `pattern` | Espera URL conter pattern |
| `wait_selector` | `selector` | Espera elemento aparecer |
| `wait_load` / `wait_navigation` | - | Espera carregamento |
| `evaluate_js` | `script` | Executa JavaScript |
| `scroll_to` | `selector` | Scrolla até elemento |
| `screenshot` | `path` (opcional) | Captura screenshot |

### Assertions

| Type | Parâmetros | Descrição |
|------|------------|-----------|
| `url_contains` | `path` | URL contém texto |
| `url_match` | `pattern` | URL match com regex |
| `title` | `expected` | Título exato da página |
| `selector_visible` | `selector` | Elemento visível |
| `selector_hidden` | `selector` | Elemento oculto/não existe |
| `element_exists` | `selector` | Elemento existe no DOM |
| `contains_text` | `expected` | Texto existe na página |
| `local_storage` | `key`, `expected` (opcional) | Valor em localStorage |
| `session_storage` | `key` | Valor em sessionStorage |
| `cookies` | `name`, `expected` (opcional) | Cookie existe |
| `wait_function` | `script` | JS retorna true |

### Teclas suportadas em `press`

`Enter`, `Escape` / `Esc`, `Tab`, `Backspace`, `Delete`, `ArrowUp`, `ArrowDown`, `ArrowLeft`, `ArrowRight`, `Space`

### Seletores CSS

Suporta seletores CSS básicos: `#id`, `.classe`, `tag`, `tag.classe`, `input[name='email']`

### Documentação completa

Para exemplos detalhados e troubleshooting, consulte:
- [docs/browser_testing.md](docs/browser_testing.md) - Documentação completa
- [examples/browser/README.md](examples/browser/README.md) - Exemplos práticos

## Configuration Files

Owl supports configuration files (`.owl`) for reusable values:

```bash
# .owl
api_base_url=https://api.example.com
api_token=Bearer my_token
```

```yaml
# test.yaml
request:
  url: "${api_base_url}/users"
  headers:
    Authorization: "${api_token}"
```

**Features:**
- Directory scoping: configs in subdirectories override parent configs
- Placeholder syntax: `${key_name}`
- Automatic loading from test file's directory tree

Consulte [docs/configuration.md](docs/configuration.md) para detalhes completos.

### Output示例

```
Found 1 browser test file(s)

Login Flow Test
========================
  ✓ Fill email field
  ✓ Fill password field
  ✓ Submit form
  ✓ Wait for redirect
  ✓ URL contains dashboard
  ✓ Title is correct
  ✓ Token saved

================================
Browser Test Summary: 1/1 passed
================================
```

## License

MIT
