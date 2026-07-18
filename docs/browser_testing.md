# Browser Testing

Owl CLI suporta testes de browser automation usando [go-rod](https://github.com/go-rod/rod), uma biblioteca Go para Chrome/Edge automation.

## Pré-requisitos

- Chrome ou Chromium instalado no sistema
- Servidor web local para servir ficheiros HTML estáticos

## Quick Start

### 1. Iniciar servidor local

```bash
npx serve examples/browser -p 8000
```

### 2. Executar testes

```bash
# Executar todos os testes de browser
go run . run examples/browser/

# Executar teste específico
go run . run examples/browser/login_test.yaml

# Com verbose para ver mais detalhes
go run . run examples/browser/login_test.yaml --verbose
```

## Estrutura de um Teste Browser

```yaml
version: 1

metadata:
  name: "Nome do Teste"
  description: "Descrição do teste"
  tags: [tag1, tag2]

browser:
  url: "http://localhost:8000/pagina.html"
  headless: true              # Executar sem interface gráfica
  timeout_seconds: 30         # Timeout global do browser
  wait_until: "networkidle2"   # Estratégia de espera
  viewport:
    width: 1280
    height: 720
  # user_agent: "custom-user-agent"  # Opcional
  # proxy: "http://proxy:8080"       # Opcional

steps:
  - name: "Descrição do passo"
    action: "nome_da_action"
    selector: "css-selector"  # varies by action
    value: "valor"            # varies by action
    timeout_seconds: 10

assertions:
  - name: "Descrição da assertion"
    type: "tipo_da_assertion"
    # parâmetros específicos
```

## Configuração do Browser

### Opções disponíveis

| Campo | Tipo | Default | Descrição |
|-------|------|---------|-----------|
| `url` | string | - | URL inicial para abrir |
| `headless` | bool | `true` | Executar sem interface gráfica |
| `timeout_seconds` | int | `30` | Timeout global |
| `wait_until` | string | `networkidle2` | Estratégia de espera |
| `viewport.width` | int | `1280` | Largura da janela |
| `viewport.height` | int | `720` | Altura da janela |
| `user_agent` | string | - | User agent customizado |
| `proxy` | string | - | Proxy HTTP |

### wait_until options

- `load` - Quando o documento principal carregar
- `domcontentloaded` - Quando o DOM estiver pronto
- `networkidle2` - Quando não houver mais de 2 conexões de rede ativas

## Actions (Steps)

As actions definem os passos da automação do browser.

### Navegação

#### `goto`
Navega para uma URL específica.

```yaml
- name: "Navegar para página de login"
  action: "goto"
  url: "http://localhost:8000/login.html"
  timeout_seconds: 15
```

#### `wait_navigation`
Espera que a navegação atual termine.

```yaml
- name: "Esperar navegação"
  action: "wait_navigation"
```

#### `wait_load`
Espera que a página carregue completamente.

```yaml
- name: "Esperar carregamento"
  action: "wait_load"
```

### Interação com Elementos

#### `click`
Clica num elemento.

```yaml
- name: "Clicar no botão de submit"
  action: "click"
  selector: "#submit-btn"
  timeout_seconds: 5
```

#### `fill` / `fill_text`
Preenche um campo de texto.

```yaml
- name: "Preencher email"
  action: "fill"
  selector: "input[name='email']"
  value: "test@example.com"
  timeout_seconds: 5

# fill_text é alias para fill
- name: "Preencher senha"
  action: "fill_text"
  selector: "input[name='password']"
  value: "minha-senha"
```

#### `hover`
Move o cursor sobre um elemento.

```yaml
- name: "Passar mouse sobre menu"
  action: "hover"
  selector: ".dropdown-menu"
```

#### `press`
Pressiona uma tecla do teclado. O elemento é focado antes de pressionar a tecla.

```yaml
- name: "Pressionar Enter"
  action: "press"
  selector: "#search-input"
  key: "Enter"
  timeout_seconds: 5
```

**Teclas suportadas:**
- `Enter`, `Escape` / `Esc`, `Tab`
- `Backspace`, `Delete`
- `ArrowUp`, `ArrowDown`, `ArrowLeft`, `ArrowRight`
- `Space`

#### `select`
Seleciona uma opção num dropdown.

```yaml
- name: "Selecionar país"
  action: "select"
  selector: "select[name='country']"
  value: "PT"
```

#### `scroll_to`
Scrolla até um elemento ficar visível.

```yaml
- name: "Scrollar até footer"
  action: "scroll_to"
  selector: "footer"
```

### Esperas

#### `wait_selector`
Espera que um elemento apareça no DOM.

```yaml
- name: "Esperar modal aparecer"
  action: "wait_selector"
  selector: ".modal"
  timeout_seconds: 10
```

#### `wait_url`
Espera que a URL mude e contenha um padrão.

```yaml
- name: "Esperar redirecionamento"
  action: "wait_url"
  pattern: "dashboard"
  timeout_seconds: 15
```

**Nota:** Este step espera que a URL mude E contenha o pattern.

### JavaScript

#### `evaluate_js`
Executa JavaScript na página e captura o resultado.

```yaml
- name: "Obter valor de variável JS"
  action: "evaluate_js"
  script: "return window.myAppState.user.name"
  timeout_seconds: 5
```

### Screenshots

#### `screenshot`
Captura screenshot da página atual.

```yaml
- name: "Capturar screenshot"
  action: "screenshot"
  path: "./screenshots/resultado.png"

# Sem path usa nome automático
- name: "Capturar screenshot automático"
  action: "screenshot"
```

## Assertions

As assertions verificam o estado final ou durante o teste.

### URL e Navegação

#### `url_contains`
Verifica se a URL contém um texto.

```yaml
- name: "URL contém dashboard"
  type: "url_contains"
  path: "dashboard"
```

#### `url_match`
Verifica se a URL matcha um padrão regex.

```yaml
- name: "URL matcha padrão"
  type: "url_match"
  pattern: "/users/\\d+/profile"
```

### Título e Conteúdo

#### `title`
Verifica o título exato da página.

```yaml
- name: "Título correto"
  type: "title"
  expected: "Dashboard | Example App"
```

#### `contains_text`
Verifica se texto existe no HTML da página.

```yaml
- name: "Mensagem de boas-vindas"
  type: "contains_text"
  expected: "Bem-vindo"
```

### Elementos DOM

#### `selector_visible`
Verifica se um elemento é visível (existe no HTML).

```yaml
- name: "Botão visível"
  type: "selector_visible"
  selector: "#submit-btn"
```

#### `selector_hidden`
Verifica se um elemento está oculto ou não existe.

```yaml
- name: "Modal fechou"
  type: "selector_hidden"
  selector: ".modal"
```

#### `element_exists`
Verifica se um elemento existe no DOM.

```yaml
- name: "Avatar carregado"
  type: "element_exists"
  selector: "img.user-avatar"
```

### Seletores CSS Suportados

| Tipo | Exemplo | Descrição |
|------|---------|-----------|
| ID | `#meu-id` | Elemento com id |
| Classe | `.minha-classe` | Elemento com classe |
| Tag | `input` | Elemento pelo nome da tag |
| Composto | `img.avatar` | Tag com classe |
| Atributo | `input[name='email']` | Input com name |

### Storage

#### `local_storage`
Verifica valores no localStorage do browser.

```yaml
# Verificar se token existe
- name: "Token guardado"
  type: "local_storage"
  key: "auth_token"

# Verificar valor específico
- name: "Email do utilizador"
  type: "local_storage"
  key: "user_email"
  expected: "test@example.com"
```

**Nota:** Para `auth_token`, a verificação é feita indiretamente através do email do utilizador.

#### `session_storage`
Verifica valores no sessionStorage.

```yaml
- name: "Dados da sessão"
  type: "session_storage"
  key: "form_data"
```

**Nota:** Verifica a existência de indicadores de sucesso na página.

### Cookies

#### `cookies`
Verifica existência e valor de cookies.

```yaml
# Verificar se cookie existe
- name: "Cookie de sessão"
  type: "cookies"
  name: "session_id"

# Verificar valor do cookie
- name: "Cookie com valor correto"
  type: "cookies"
  name: "user_prefs"
  expected: "dark_mode"
```

### JavaScript

#### `wait_function`
Executa uma função JavaScript e verifica se retorna `true`.

```yaml
- name: "App inicializado"
  type: "wait_function"
  script: "window.app && window.app.isReady()"
```

## Exemplos Completos

### Login Flow

```yaml
version: 1

metadata:
  name: "Login Flow Test"
  tags: [login, smoke]

browser:
  url: "http://localhost:8000/login.html"
  headless: true
  timeout_seconds: 30
  wait_until: "networkidle2"
  viewport:
    width: 1280
    height: 720

steps:
  - name: "Preencher email"
    action: "fill"
    selector: "input[name='email']"
    value: "test@example.com"

  - name: "Preencher senha"
    action: "fill"
    selector: "input[name='password']"
    value: "KFwBNLfFXn71rCb5QJuU"

  - name: "Clicar submit"
    action: "click"
    selector: "#submit-btn"

  - name: "Esperar dashboard"
    action: "wait_url"
    pattern: "dashboard"
    timeout_seconds: 15

assertions:
  - name: "URL contém dashboard"
    type: "url_contains"
    path: "dashboard"

  - name: "Título correto"
    type: "title"
    expected: "Dashboard | Example App"

  - name: "Token guardado"
    type: "local_storage"
    key: "auth_token"
```

### Search Flow

```yaml
version: 1

metadata:
  name: "Search Test"
  tags: [search]

browser:
  url: "http://localhost:8000/search.html"
  headless: true
  timeout_seconds: 30

steps:
  - name: "Navegar para busca"
    action: "goto"
    url: "http://localhost:8000/search.html"

  - name: "Clicar ícone de busca"
    action: "click"
    selector: ".search-icon"

  - name: "Esperar input"
    action: "wait_selector"
    selector: "#search-input"

  - name: "Digitar consulta"
    action: "fill_text"
    selector: "#search-input"
    value: "golang tutorial"

  - name: "Submeter com Enter"
    action: "press"
    selector: "#search-input"
    key: "Enter"

  - name: "Esperar resultados"
    action: "wait_selector"
    selector: ".search-results"

assertions:
  - name: "URL contém query"
    type: "url_contains"
    path: "#search?q=golang"

  - name: "Resultados visíveis"
    type: "selector_visible"
    selector: ".search-results"

  - name: "Contém termo buscado"
    type: "contains_text"
    expected: "golang"
```

### Form Submission

```yaml
version: 1

metadata:
  name: "Contact Form Test"
  tags: [form]

browser:
  url: "http://localhost:8000/form.html"
  headless: true
  timeout_seconds: 30

steps:
  - name: "Preencher nome"
    action: "fill"
    selector: "input[name='name']"
    value: "John Doe"

  - name: "Preencher email"
    action: "fill"
    selector: "input[name='email']"
    value: "john@example.com"

  - name: "Preencher assunto"
    action: "fill"
    selector: "input[name='subject']"
    value: "Test Subject"

  - name: "Preencher mensagem"
    action: "fill"
    selector: "textarea[name='message']"
    value: "This is a test message."

  - name: "Submeter"
    action: "click"
    selector: "button[type='submit']"

  - name: "Esperar sucesso"
    action: "wait_selector"
    selector: ".success-message"

assertions:
  - name: "Mensagem de sucesso"
    type: "selector_visible"
    selector: ".success-message"

  - name: "Dados na sessão"
    type: "session_storage"
    key: "last_form_submission"

  - name: "Contém agradecimento"
    type: "contains_text"
    expected: "Thank you"

  - name: "Sem erros"
    type: "selector_hidden"
    selector: ".error-message"
```

## Debugging

### Verbose Mode

Use `--verbose` ou `-v` para ver mais detalhes:

```bash
go run . run examples/browser/login_test.yaml --verbose
```

### Screenshots on Failure

```bash
go run . run examples/browser/login_test.yaml --screenshot ./screenshots/
```

### wait_url não funciona

Se `wait_url` não funciona, verifique:

1. A URL muda realmente durante o teste?
2. O pattern especificado está correto?
3. O timeout é suficiente?

Alternativa: use `wait_selector` para esperar um elemento que aparece após a navegação.

### press não funciona

Se `press Enter` não funciona:

1. Confirme que o elemento está focado
2. Verifique se o JavaScript da página escuta o evento correto (`keydown`, `keyup`, ou `keypress`)
3. Considere usar `click` num botão de submit se disponível

### Assertions de storage falham

- `local_storage` e `session_storage` verificam estado após execução do JavaScript
- Para `local_storage`, o `auth_token` é verificado indiretamente via presença do email
- Verifique se o JavaScript da página define os valores corretamente
