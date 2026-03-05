# Template Engine Reference

Este documento é a referência canônica do sistema de interpolação de templates do Trellis. Aplica-se à v0.7.16+.

## 1. O que é o Interpolator

O `Interpolator` é um **port funcional** injetável no engine:

```go
type Interpolator func(ctx context.Context, templateStr string, data any) (string, error)
```

Toda função compatível com esse tipo pode ser injetada via `NewEngine`:

```go
engine := runtime.NewEngine(loader, nil, runtime.HTMLInterpolator)
```

Se nenhum interpolador for fornecido, `DefaultInterpolator` é usado automaticamente.

## 2. Variantes Disponíveis

| Variante | Template Package | Escape HTML? | Quando usar |
|:---|:---|:---|:---|
| `DefaultInterpolator` | `text/template` | ❌ Não | CLI, texto puro, Markdown |
| `HTMLInterpolator` | `html/template` | ✅ Sim | Chat UI, SSE, output direto ao browser |
| `LegacyInterpolator` | `strings.ReplaceAll` | ❌ Não | Flows legados com `{{ key }}` sem ponto |

> [!NOTE]
> `DefaultInterpolator` é o padrão. Para o Chat UI embutido (`trellis serve`), considere injetar `HTMLInterpolator` para evitar XSS em conteúdo dinâmico.

## 3. Contexto de Template

Dentro dos templates, as seguintes variáveis estão disponíveis:

| Expressão | Origem | Exemplo |
|:---|:---|:---|
| `{{ .key }}` | `state.Context` (user space) | `{{ .username }}` |
| `{{ .sys.ans }}` | `state.SystemContext` | Último input do usuário |
| `{{ .sys.* }}` | `state.SystemContext` | Namespace do sistema (somente leitura) |
| `{{ .tool_result.id }}` | Auto-injetado após tool (v0.7.16+) | ID do último tool call |
| `{{ .tool_result.result }}` | Auto-injetado após tool (v0.7.16+) | Resultado do último tool call |

### Chaves Reservadas

| Chave | Descrição |
|:---|:---|
| `sys.*` | Namespace do sistema. Somente leitura nos templates. Protegido contra gravação via `save_to`. |
| `tool_result` | Último resultado de ferramenta bem-sucedido (política **last-result**, v0.7.16+). |
| `tool_results` | **Reservado** para política de acumulação futura (v0.8+). Não use em flows hoje. |

## 4. FuncMap — Funções Disponíveis

### Funções Registradas (v0.7.16+)

| Função | Uso | Comportamento |
|:---|:---|:---|
| `default` | `{{ default "N/A" .key }}` | Retorna `.key` se não-zero; senão, retorna `"N/A"` |
| `coalesce` | `{{ coalesce .a .b .c }}` | Retorna o primeiro valor não-zero da lista |
| `toJson` | `{{ toJson .obj }}` | Serializa para JSON; propaga erros de marshal |

### Funções Nativas do Go Template

| Função | Exemplo |
|:---|:---|
| `index` | `{{ index .config "env" }}` — acessa mapa dinâmico |
| `if` / `else` | `{{ if .logged_in }}...{{ end }}` |
| `eq`, `ne`, `lt`, `gt` | `{{ if eq .status "ok" }}...{{ end }}` |
| `range` | `{{ range .items }}{{ . }}{{ end }}` |
| `len` | `{{ len .items }}` |

## 5. Comportamento de Chave Ausente

```
{{ .missing_key }}  →  <no value>
```

Chaves ausentes em `map[string]any` sempre rendem `<no value>` em Go templates. Isso é comportamento padrão do `text/template` e `html/template` e **não é configurável** para maps.

Use `{{ default }}` para fornecer valores explícitos:

```
{{ default "visitante" .username }}  →  "visitante"  (se .username ausente ou vazio)
{{ default "visitante" .username }}  →  "Alice"      (se .username = "Alice")
```

## 6. Exemplos Rápidos

```markdown
# Interpolação simples
Olá, {{ .username }}!

# Com fallback
Olá, {{ default "visitante" .username }}!

# Condicional
{{ if .tool_result.result }}Ferramenta executou com sucesso.{{ else }}Aguardando.{{ end }}

# Acesso ao resultado de ferramenta (após nó tool)
- ID da chamada: {{ .tool_result.id }}
- Resultado: {{ .tool_result.result }}

# Inspecionar resultado como JSON
{{ toJson .tool_result }}

# Acesso a mapa dinâmico
{{ index .config "environment" }}

# Comparação de string
{{ if eq .user_input "sim" }}Confirmado!{{ end }}
```

## 7. Limitações Conhecidas

| Cenário | Comportamento | Mitigação |
|:---|:---|:---|
| Chave ausente em map | Rende `<no value>` | Use `{{ default "fallback" .key }}` |
| `{{ .tool_result }}` raw | Rende representação Go do map | Use `{{ .tool_result.result }}` ou `{{ toJson .tool_result }}` |
| Template inválido | Retorna erro (não silencioso) | Corrija a sintaxe do template |
| Strings com `<`, `>` no `DefaultInterpolator` | **Não** são escapadas (text/template) | Use `HTMLInterpolator` para output no browser |
