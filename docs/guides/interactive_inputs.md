# Interactive Inputs

Trellis allows flows to pause and request data from the user (or system). This is done using **Question Nodes** with specific metadata.

## Concept: Delegated UI

Trellis does **not** render widgets (like Select Boxes or Date Pickers). Instead, it declares an *Intent* for input.
The **Host** (CLI, Web App, Agent) is responsible for interpreting this intent and rendering the appropriate controls.

## Defining Inputs (Markdown)

To create an input, use a node with `type: question` and add `input_*` fields in the frontmatter.

```markdown
---
id: favorite_color
type: question
input_type: choice
input_options: 
  - Red
  - Blue
  - Green
input_default: Blue
transitions:
  - to: next_step
---
# Favorite Color

Please select your favorite color from the list.
```

### Supported Fields

| Field | Type | Description |
| :--- | :--- | :--- |
| `input_type` | `string` | The kind of data requested (e.g., `text`, `choice`, `confirm`). |
| `input_options` | `list` | (Optional) Valid values for `choice` type. |
| `input_default` | `string` | (Optional) Default value if the user skips. |
| `input_secret` | `bool` | (Optional) if true, indicates sensitive input (passwords). |

## Using Inputs in Transitions

The value provided by the user is available in the `input` variable during transition evaluation.

```yaml
transitions:
  - to: node_red
    condition: input == "Red"
  - to: node_blue
    condition: input == "Blue"
```

## Host Implementation

When the Engine encounters a Question Node, it returns a `ActionRequestInput` action.

### Using the Runner (CLI)

The standard `trellis.Runner` already implements listeners for these actions and renders:

- `choice`: Using a list selector.
- `text`: Using a standard prompt.
- `confirm`: Using a Yes/No prompt.

### using the Low-Level API

If you are building a custom integration, you need to handle the `ActionRequestInput`:

```go
for _, act := range actions {
    if req, ok := act.Payload.(domain.InputRequest); ok {
        // Render your custom UI widget here
        // Then pass the result to engine.Navigate(ctx, state, result)
    }
}
```
