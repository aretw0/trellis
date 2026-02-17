# Typed Flow Example

This example demonstrates `context_schema` validation for typed context values.

## Run

Valid context:

```bash
trellis run ./examples/typed-flow --context '{"api_key":"secret","retries":3,"tags":["prod","critical"]}'
```

Invalid context (wrong type):

```bash
trellis run ./examples/typed-flow --context '{"api_key":"secret","retries":"three","tags":["prod"]}'
```

Missing field:

```bash
trellis run ./examples/typed-flow --context '{"api_key":"secret"}'
```
