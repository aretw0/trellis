---
default_context:
  api_url: "http://localhost:8080"
  user_role: "guest"
  retries: 3
  debug_mode: true
to: show_config
---
# Default Context Demo

This flow demonstrates how `default_context` works.

We have defined defaults in this file:

- API URL: `http://localhost:8080`
- Role: `guest`
- Retries: `3`

You can override these by running:

```bash
trellis run examples/default-context-demo --context '{"user_role": "admin"}'
```
