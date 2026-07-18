# Browser Test Examples

Este diretório contém exemplos de testes de browser automation para o Owl CLI, usando [go-rod](https://github.com/go-rod/rod).

## Quick Start

### 1. Iniciar servidor local

```bash
npx serve . -p 8000
```

### 2. Executar testes

```bash
# Todos os testes de browser
go run . run examples/browser/

# Teste específico
go run . run examples/browser/login_test.yaml

# Com verbose para ver mais detalhes
go run . run examples/browser/login_test.yaml --verbose
```

## Ficheiros

| Ficheiro | Descrição |
|----------|-----------|
| `login.html` | Página de login com validação JavaScript |
| `login_test.yaml` | Teste completo do fluxo de login |
| `dashboard.html` | Página do dashboard pós-login |
| `form.html` | Página de formulário de contacto |
| `form_test.yaml` | Teste de preenchimento de formulário |
| `search.html` | Página de busca com modal |
| `search_test.yaml` | Teste de pesquisa com hash URL |

## Credenciais de Teste

- **Email**: `test@example.com`
- **Senha**: `KFwBNLfFXn71rCb5QJuU`

## Exemplos

### 1. Login Flow (`login_test.yaml`)

Testa o fluxo completo de autenticação:
- Preenchimento de email e senha
- Submissão do formulário
- Verificação de redirecionamento para dashboard
- Validação de localStorage após login

**Assertions verificadas:**
- URL contém "dashboard"
- Título da página correto
- Mensagem de boas-vindas visível
- Token de autenticação guardado em localStorage

### 2. Search Flow (`search_test.yaml`)

Testa a funcionalidade de pesquisa:
- Abertura do modal de busca
- Preenchimento do campo de pesquisa
- Submissão com tecla Enter
- Verificação do hash na URL

**Assertions verificadas:**
- URL contém hash `#search?q=golang`
- Resultados de pesquisa visíveis
- Página contém termo procurado

### 3. Form Submission (`form_test.yaml`)

Testa o preenchimento de um formulário de contacto:
- Preenchimento de múltiplos campos
- Submissão do formulário
- Verificação de mensagem de sucesso

**Assertions verificadas:**
- Mensagem de sucesso visível
- Dados guardados em sessionStorage
- Texto de confirmação presente
- Sem mensagens de erro

## Estrutura de um Teste Browser

```yaml
version: 1

metadata:
  name: "Nome do Teste"
  description: "Descrição detalhada"
  tags: [tag1, tag2]

browser:
  url: "http://localhost:8000/pagina.html"
  headless: true
  timeout_seconds: 30
  wait_until: "networkidle2"
  viewport:
    width: 1280
    height: 720

steps:
  - name: "Descrição do passo"
    action: "fill|click|press|..."
    selector: "css-selector"
    value: "valor"
    timeout_seconds: 10

assertions:
  - name: "Descrição da assertion"
    type: "url_contains|selector_visible|..."
    selector: "css-selector"
    expected: "valor esperado"
```

## Actions Disponíveis

| Action | Parâmetros | Descrição |
|--------|------------|-----------|
| `goto` | `url` | Navega para URL |
| `click` | `selector` | Clica num elemento |
| `fill` / `fill_text` | `selector`, `value` | Preenche campo de texto |
| `hover` | `selector` | Move mouse sobre elemento |
| `select` | `selector`, `value` | Seleciona opção em dropdown |
| `press` | `selector`, `key` | Pressiona tecla (Enter, Escape, etc.) |
| `wait_url` | `pattern` | Espera URL conter pattern |
| `wait_selector` | `selector` | Espera elemento aparecer |
| `wait_load` | - | Espera página carregar |
| `wait_navigation` | - | Espera navegação completar |
| `evaluate_js` | `script` | Executa JavaScript |
| `scroll_to` | `selector` | Scrolla até elemento |
| `screenshot` | `path` (opcional) | Captura screenshot |

## Assertions Disponíveis

| Type | Parâmetros | Descrição |
|------|------------|-----------|
| `url_contains` | `path` | Verifica URL contém texto |
| `url_match` | `pattern` | Verifica URL com regex |
| `title` | `expected` | Verifica título exato |
| `selector_visible` | `selector` | Elemento é visível |
| `selector_hidden` | `selector` | Elemento oculto/não existe |
| `element_exists` | `selector` | Elemento existe no DOM |
| `contains_text` | `expected` | Texto existe na página |
| `local_storage` | `key`, `expected` (opcional) | Verifica localStorage |
| `session_storage` | `key` | Verifica sessionStorage |
| `cookies` | `name`, `expected` (opcional) | Verifica cookies |
| `wait_function` | `script` | JS retorna true |

## Teclas Suportadas em `press`

- `Enter`
- `Escape` / `Esc`
- `Tab`
- `Backspace`
- `Delete`
- `ArrowUp` / `ArrowDown` / `ArrowLeft` / `ArrowRight`
- `Space`

## Seletores CSS

Suporta todos os seletores CSS básicos:

- `#id` - Elemento por ID
- `.classe` - Elemento por classe
- `tag` - Elemento por nome de tag
- `tag.classe` - Tag com classe
- `input[name='email']` - Input com atributo name

## Notas Importantes

1. **`wait_url`** espera que a URL mude E contenha o pattern. Se a URL já contém o pattern no início, pode falhar.

2. **Assertions `selector_visible` e `element_exists`** usam parsing de HTML, não JavaScript. Verificam se o elemento existe no source da página.

3. **`local_storage` para `auth_token`** é verificado indiretamente: se o email do utilizador aparece na página, significa que o login teve sucesso e o token foi guardado.

4. **`press Enter`** foca o elemento primeiro, depois pressiona a tecla usando `page.Keyboard.Press()`. O elemento precisa estar visível e pronto para receber eventos de teclado.

5. **Screenshots** são salvos apenas quando configurado com `--screenshot <path>` ou quando o teste falha em modo verbose.

## Debugging

```bash
# Verbose mode
go run . run examples/browser/login_test.yaml --verbose

# Com screenshots
go run . run examples/browser/login_test.yaml --screenshot ./debug/

# Executar apenas um teste
go run . run examples/browser/login_test.yaml
```
