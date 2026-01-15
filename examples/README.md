# Trellis Examples

This directory contains examples demonstrating different ways to use the Trellis engine.

## 1. Hello World (Standard)

**Path:** [`hello-world/`](./hello-world)
**Concepts:** `MemoryLoader`, `TUI`, `trellis.Runner`
The standard entry point for developers building internal tools. It shows how to define a graph in Go code (in-memory) and run it using the standard Runner.

## 2. Low Level API (Advanced)

**Path:** [`low-level-api/`](./low-level-api)
**Concepts:** `Manual Loop`, `engine.Render`, `engine.Navigate`
Demonstrates how to manually drive the engine without using the `trellis.Runner`. Useful if you need to integrate Trellis into a custom UI framework, a game engine, or a non-standard event loop.

## 3. Tour (The Product)

**Path:** [`tour/`](./tour)
**Concepts:** `Loam Adapter`, `Markdown Files`, `CLI`
A content-heavy example that demonstrates the features of the Trellis file format (`.md` files). This is what you run with `trellis run ./examples/tour`.

## 4. Observability (Hooks & Debug)

**Path:** [`observability/`](./observability)
**Concepts:** `LifecycleHooks`, `--debug`, `Events`
Demonstrates how to use the `--debug` flag to visualize state transitions and events in the console.

## 5. Structured Logging (Production)

**Path:** [`structured-logging/`](./structured-logging)
**Concepts:** `slog` (JSON Logs), `Prometheus` (Metrics)
Demonstrates industry-standard observability by integrating Trellis with Go's `log/slog` and `prometheus/client_golang`. Shows how to output machine-readable logs and metrics.

## 6. Fail Fast (Data Validation)

**Path:** [`fail_fast/`](./fail_fast)
**Concepts:** `required_context`, `Fail Fast`
Demonstrates how to use `required_context` to enforce data dependencies at the start of a flow. This shows how Trellis protects against missing execution context by failing immediately with a clear error.
