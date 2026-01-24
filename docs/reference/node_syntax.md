# The Trellis Guide: Node Syntax Reference

This document is the canonical reference for defining Nodes in Trellis. It reflects the **v0.7+ Universal Action Semantics** ("Duck Typing").

## 1. Philosophies

### 1.1. Everything is a Node

The Node is the atomic unit of the conversation/flow.

- It can **Speak** (Text).
- It can **Listen** (Input).
- It can **Act** (Tool).
- It can **Decide** (Transition).

### 1.2. Behavioral Typing (Duck Typing)

We do not rely on rigid `type` fields (like `type: tool`). Instead, the properties you define determine the node's behavior.

- **Has `do`?** -> It is an **Action Node**. The Engine executes the tool.
- **Has `wait` or `input_type`?** -> It is an **Input Node**. The Engine pauses for user input.
- **Has `content`?** -> It renders text (Markdown).

> **Note**: You can mix behaviors, with constraints.
>
> - `Content` + `Do` = **"Talk & Act"** (e.g. "Loading..." + Init).
> - `Content` + `Wait` = **"Talk & Listen"** (e.g. "Question?" + Input).
> - **Forbidden**: `Do` + `Wait` (in the same node). You cannot act and listen simultaneously (state collision).

## 2. Anatomy of a Node (YAML/Frontmatter)

```yaml
id: my_node_id          # Optional (Defaults to filename)
type: text              # Legacy/Optional. Inferred from behavior.

# --- Behavior: Action (The "Do") ---
do:
  name: my_tool_name    # Tool to execute
  args:                 # Arguments passed to the tool
    id: "{{ user_id }}"

# --- Behavior: Input (The "Wait") ---
wait: true              # Pauses for simple text input (Enter)
# OR
input_type: confirm     # Pauses for typed input (e.g., [y/N])
input_options: ["A","B"]
save_to: my_variable    # Saves input (or tool result) to Context

# --- Flow Control ---
to: next_node_id        # Unconditional transition
# OR
transitions:
  - condition: input == 'success'
    to: success_node
  - to: fallback_node

on_error: error_handler_node  # Transition if Tool fails
on_signal:
  timeout: timeout_handler    # Transition on signal
```

## 3. Formatting Rules

### 3.1. Markdown Body = Content

In Markdown files, any text **below** the frontmatter is the `content`.

```markdown
---
do: init_db
to: menu
---
**Initializing Database...**
Please wait while we set up tables.
```

### 3.2. JSON/YAML = Explicit Content

In strictly structured files, use the `content` key.

```yaml
id: start
do: 
  name: init_db
content: "Initializing Database..."
to: menu
```

## 4. Universal Action Patterns

### 4.1. Text + Action (The "Zero Fatigue" Pattern)

Render a message and immediately execute a backend task. The transition happens when the task receives a Result.

```yaml
# Loading Screen
content: "Checking credentials..."
do: check_creds
save_to: auth_result
transitions:
  - condition: input.is_valid
    to: dashboard
  - to: login_fail
```

**How it works:**

1. Engine renders "Checking credentials...".
2. Engine executes `check_creds`.
3. Tool returns result `{"is_valid": true}`.
4. Result is saved to `auth_result`.
5. Transition condition `input.is_valid` is evaluated (True).
6. Engine moves to `dashboard`.

### 4.2. Explicit Questions & Options

For interactions that require specific choices, use `type: question` (optional if behaviors are clear) or `input_type: choice`.

```yaml
# Simple Yes/No
content: "Proceed?"
input_type: confirm
save_to: proceed

# Multiple Choice (Syntactic Sugar)
content: "Pick a color:"
options:
  - "Red"
  - "Blue"
transitions:
  - condition: input == 'Red'
    to: red_pill
```

> **Note on Options**: The `options` list implies `input_type: choice`. The engine presents these to the user (e.g. arrow keys in CLI).

### 4.3. Action Safety (Error Handling)

If an action fails, `on_error` takes precedence over `to`.

```yaml
do: critical_op
on_error: rollback_node
to: success_node
```

- **Success**: Goes to `success_node`.
- **Failure**: Goes to `rollback_node`.

### 4.4. Scriptable Tools (v0.7+)

Define ad-hoc scripts inline (requires `--unsafe-inline`) or via `tools.yaml`.

```yaml
# Ad-hoc execution (Dev Mode)
do:
  name: quick_script
  x-exec:
    command: python
    args: ["scripts/calc.py"]
```

## 5. Property Dictionary

| Property | Type | Description |
| :--- | :--- | :--- |
| `id` | `string` | Unique identifier. Defaults to filename (without ext). |
| `do` | `ToolCall` | Definition of side-effect to execute. |
| `wait` | `bool` | If true, pause for user input (default text). |
| `content` | `string` | Message to display to the user. |
| `options` | `[]string` | Shorthand for choice input. Presents a menu. |
| `input_type`| `string` | `text` (default), `confirm`, `choice`, `int`. |
| `input_default`| `string` | Default value if user presses Enter. |
| `input_options`| `[]string` | Options for `choice` input (Low-level). |
| `save_to` | `string` | Context variable key to store Input or Tool Result. |
| `to` | `string` | Shorthand for single unconditional transition. |
| `transitions` | `[]Transition` | List of conditional paths. Evaluated in order. |
| `on_error` | `string` | Target node ID if `do` fails. |
| `on_signal` | `map[string]string` | Handlers for global signals (`interrupt`, `timeout`). |
| `tools` | `[]Tool` | Definitions of tools available to this node (for LLMs). |

### 5.1. The Confirm Convention (Unix Style)

When using `input_type: confirm`, Trellis follows the standard CLI convention:

- **Empty Input (Enter)**: Defaults to `yes` (True) unless an explicit `input_default` is provided.
- **Strict Validation**: Only `y`, `yes`, `true`, `1` (True) or `n`, `no`, `false`, `0` (False) are accepted.
- **Normalization**: Input is automatically converted to a canonical `yes` or `no` before being saved to `save_to` or evaluated in `transitions`.

**Implementation Example:**

```yaml
input_type: confirm
input_default: "no" # Overrides convention to make Enter = False
on_denied: stop_flow
to: continue_flow
```
