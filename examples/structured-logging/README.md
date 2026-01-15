# Structured Logging & Prometheus Example

This example demonstrates how to implement **Industry Standard Observability** in Trellis applications using:

1. **`log/slog`**: Go's standard library for structured JSON logging.
2. **`prometheus`**: Metrics collection for monitoring.
3. **Graceful Shutdown**: Proper handling of `SIGINT` (Ctrl+C) to ensure cleanup.

## Usage

Run the example:

```bash
go run ./examples/structured-logging
```

The flow will start and log events to `stdout` in JSON format.

## Cancellation

You can cancel the execution at any time by pressing **Ctrl+C**. The application handles the intersection signal, cancels the context, and shuts down gracefully.

```bash
# JSON Output
{"time":"...","level":"INFO","msg":"received interrupt signal, shutting down..."}
```

## Verifying Metrics

The application exposes a Prometheus metrics endpoint on port `:2112`.

While the application is running (or after it finishes, as the server stays alive until a second Ctrl+C), check the metrics:

```bash
curl http://localhost:2112/metrics | grep trellis
```

**Expected Output:**

```text
# HELP trellis_node_visits_total Total number of node visits
# TYPE trellis_node_visits_total counter
trellis_node_visits_total{node_id="start"} 1
```
