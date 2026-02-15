# Introspection Example

This example demonstrates **real-time state observation** using Trellis's introspection capabilities.

## Purpose

Shows how to monitor a running Trellis flow's state using the `introspection` library, useful for:

- **Dashboards**: Building real-time UI monitoring
- **Metrics**: Collecting execution statistics
- **Debugging**: Observing state transitions live

## Key Concepts

### `State()` vs `Watch()`

- **`State()`**: Poll-based snapshot (pull model)
  - Call anytime to get current state
  - Synchronous
  - Ideal for HTTP endpoints (`GET /state`)

- **`Watch(ctx)`**: Stream-based changes (push model)
  - Receive events when state changes
  - Asynchronous channel
  - Ideal for real-time feeds (WebSocket, SSE)

### Thread Safety

The `Runner` implements `introspection.TypedWatcher[*domain.State]`:

- All snapshots are cloned (deep copy)
- Safe for concurrent access
- No race conditions

## Running

```bash
cd examples/observability-introspection
go run main.go
```

**Expected Behavior**:

1. Observer goroutine prints state changes every transition
2. Interactive flow runs in foreground
3. Press Ctrl+C to gracefully exit both

## Architecture

```
┌─────────────┐
│   Runner    │ (Execution)
│             │
│  Watch(ctx) │──┐
└─────────────┘  │
                 ├──> StateChange Stream
┌─────────────┐  │
│ Aggregator  │<─┘
│             │
│  Watch(ctx) │───> Observer Goroutine
└─────────────┘         │
                        ▼
                  Console Output
```

## Differences from Hooks

| Feature | Hooks (OnNodeEnter) | Introspection (Watch) |
|---------|---------------------|----------------------|
| Model | Push (events) | Push (state changes) |
| Granularity | Per-event | Per-state snapshot |
| Use Case | Logging, tracing | Dashboards, metrics |
| Blocking | Can block engine | Never blocks |

## Production Notes

- **Backpressure**: Slow consumers drop events (non-blocking)
- **Memory**: Each watcher has a 10-event buffer
- **Cleanup**: Watchers auto-close on context cancel
- **Stdin Goroutine**: Properly cancels on `ctx.Done()` (no leaks)

See `examples/structured-logging` for hook-based observability.
