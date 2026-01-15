# Fail Fast Validation Example

This example demonstrates the **Data Contract** feature of Trellis.

The `start.md` node defines a `required_context` list. This ensures that the flow **fails immediately** if the necessary context variables (`api_key`, `user_id`) are missing.

## How to Run

### 1. Run (Expected Failure)

Since the CLI currently does not inject context by default, running this will fail:

```bash
go run ./cmd/trellis run ./examples/fail_fast
```

**Expected Output:**

```text
Error: Node 'start' requires context keys that are missing: [api_key user_id]
```

### 2. Simulating Success

To see it pass, typically the calling application (Host) would inject the context. In the future, the CLI will support a `--context` flag.

For now, you can verify this behavior by creating a small Go program that seeds the context, or by observing the error message above which confirms the validation logic is active.
