---
to: end_node
---
# Kitchen Sink

## String Interpolation (via save_to)

- user_input: {{ .user_input }}

## Conditional — non-null value is truthy
{{ if .user_input }}user_input is set{{ else }}user_input is not set{{ end }}

## String comparison with eq
{{ if eq .user_input "Hello Tool" }}user_input matches Hello Tool{{ else }}user_input does not match{{ end }}

## Tool Result (Typed access — Schema: {{ .tool_result._id }})

- tool_result ID: {{ .tool_result._id }}
- tool_result field: {{ .tool_result.received }}
- tool_result raw: {{ .tool_result }}
