---
to: end_node
---
# Kitchen Sink

## Interpolação de string (via save_to)

- user_input: {{ .user_input }}

## Condicional — valor não-nulo é truthy
{{ if .user_input }}user_input is set{{ else }}user_input is not set{{ end }}

## Comparação de string com eq
{{ if eq .user_input "Hello Tool" }}user_input matches Hello Tool{{ else }}user_input does not match{{ end }}

## Resultado de tool (mapa Go bruto — estrutura: map[id result])

- tool_result raw: {{ .tool_result }}
