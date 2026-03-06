# Template Engine Reference

This document is the canonical reference for the Trellis template interpolation system (v0.7.16+).

## 1. What is the Interpolator

The `Interpolator` is a functional **port** that can be injected into the engine:

```go
type Interpolator func(ctx context.Context, templateStr string, data any) (string, error)
```

Any function compatible with this signature can be injected via `NewEngine`:

```go
engine := runtime.NewEngine(loader, nil, runtime.HTMLInterpolator)
```

If no interpolator is provided, `DefaultInterpolator` is used by default.

## 2. Available Variants

| Variant | Template Package | HTML Escaping | Recommended Use |
|:---|:---|:---|:---|
| `DefaultInterpolator` | `text/template` | ❌ No | CLI, Plain Text, Markdown |
| `HTMLInterpolator` | `html/template` | ✅ Yes | Chat UI, SSE, direct browser output |
| `LegacyInterpolator` | `strings.ReplaceAll` | ❌ No | Legacy flows using `{{ key }}` without a leading dot |

> [!NOTE]
> `DefaultInterpolator` is the standard. For the built-in Chat UI (`trellis serve`), consider injecting `HTMLInterpolator` to prevent XSS.

## 3. Template Context

The following variables are available within templates:

| Expression | Origin | Example |
|:---|:---|:---|
| `{{ .key }}` | `state.Context` (user space) | `{{ .username }}` |
| `{{ .sys.ans }}` | `state.SystemContext` | Last user input |
| `{{ .sys.* }}` | `state.SystemContext` | System namespace (read-only) |
| `{{ .tool_result._id }}` | Auto-injected after tools | Call ID of the last tool call |
| `{{ .tool_result.result }}` | Auto-injected after tools | Result (if scalar) |
| `{{ .tool_result.field }}` | Auto-injected after tools | Extracted field (if result was a map) |

### Reserved Keys

| Key | Description |
|:---|:---|
| `sys.*` | System namespace. Read-only in templates. Protected from `save_to` writes. |
| `tool_result` | Last successful tool result (Policy: **last-result**, v0.7.16+). |
| `tool_results` | **Reserved** for future accumulation policy (v0.8+). Do not use in flows today. |

## 4. FuncMap — Available Functions

### Standard Functions (v0.7.16+)

| Function | Usage | Behavior |
|:---|:---|:---|
| `default` | `{{ default "N/A" .key }}` | Returns `.key` if non-zero; otherwise returns `"N/A"` |
| `coalesce` | `{{ coalesce .a .b .c }}` | Returns the first non-zero value in the list |
| `toJson` | `{{ toJson .obj }}` | Serializes to JSON; propagates marshal errors |

### Native Go Template Functions

| Function | Example |
|:---|:---|
| `index` | `{{ index .config "env" }}` — access dynamic maps |
| `if` / `else` | `{{ if .logged_in }}...{{ end }}` |
| `eq`, `ne`, `lt`, `gt` | `{{ if eq .status "ok" }}...{{ end }}` |
| `range` | `{{ range .items }}{{ . }}{{ end }}` |
| `len` | `{{ len .items }}` |

## 5. Missing Key Behavior

```
{{ .missing_key }}  →  <no value>
```

Missing keys in `map[string]any` always render as `<no value>` in Go templates. This is the default behavior of `text/template` and `html/template` and is **not configurable** for maps (unlike structs).

Use `{{ default }}` to provide explicit fallback values:

```
{{ default "guest" .username }}  →  "guest"  (if .username is missing or zero)
{{ default "guest" .username }}  →  "Alice"  (if .username = "Alice")
```

## 6. Quick Examples

```markdown
# Simple interpolation
Hello, {{ .username }}!

# With fallback
Hello, {{ default "guest" .username }}!

# Conditional
{{ if .tool_result.result }}Tool executed successfully.{{ else }}Waiting.{{ end }}

# Accessing tool results (after a tool node)
- Call ID: {{ .tool_result._id }}
- Result: {{ .tool_result.result }}

# Inspecting result as JSON
{{ toJson .tool_result }}

# Dynamic map access
{{ index .config "environment" }}

# String comparison
{{ if eq .user_input "yes" }}Confirmed!{{ end }}
```

## 7. Known Limitations

| Scenario | Behavior | Mitigation |
|:---|:---|:---|
| Missing map key | Renders `<no value>` | Use `{{ default "fallback" .key }}` |
| `{{ .tool_result }}` raw | Renders Go map representation | Use `{{ .tool_result.result }}` or `{{ toJson .tool_result }}` |
| Invalid template | Returns error (stops flow) | Ensure correct template syntax |
| HTML in `DefaultInterpolator` | **Not** escaped (text/template) | Use `HTMLInterpolator` for browser output |
