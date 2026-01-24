# Interrupt Demo

This example demonstrates how to handle **Global Signals** (specifically `SIGINT` / `Ctrl+C`) in Trellis.

## The Problem

In many CLI applications, pressing `Ctrl+C` kills the process immediately, leaving resources dangling or transactions incomplete. Trellis allows you to intercept this signal and transition to a specific node (e.g., a confirmation screen or cleanup routine) instead of crashing.

## The Solution

The `on_signal` property in the `start.md` node defines the handler:

```yaml
on_signal:
  interrupt: confirm_exit
```

When `Ctrl+C` is pressed, the engine pauses the current action and transitions to `confirm_exit`.

## Running the Demo

```bash
go run ./examples/interrupt-demo
```

### Expected Behavior

1. The flow starts and waits for input.
2. **Press `Ctrl+C`**.
3. Instead of exiting, the flow jumps to the `confirm_exit` node:
    > "You pressed Ctrl+C. Are you sure you want to exit?"
4. Type `y` (or `yes`) to exit gracefully, or `n` to return to the start.

### Windows Note

On Windows, `Ctrl+C` typically closes the standard input stream (`os.Stdin`), which can cause race conditions. Trellis automatically detects this and switches to `CONIN$` to ensure the signal is caught correctly without killing the input stream.
