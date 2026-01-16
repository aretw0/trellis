# Trellis Examples

This directory contains examples demonstrating different ways to use the Trellis engine.

## 1. Getting Started

### [Tour (The Product)](./tour)

**Concepts:** `Loam Adapter`, `Markdown Files`, `CLI`
A content-heavy example that demonstrates the features of the Trellis file format (`.md` files). This is what you run with `trellis run ./examples/tour`.

### [Hello World (Standard)](./hello-world)

**Concepts:** `MemoryLoader`, `TUI`, `trellis.Runner`
The standard entry point for developers building internal tools. It shows how to define a graph in Go code (in-memory) and run it using the standard Runner.

---

## 2. Core Features

### [Default Context (Mocking)](./default-context-demo)

**Concepts:** `default_context`, `Mocking`, `Local Dev`
Demonstrates how to define fallback values in `start.md`. These defaults act as mocks for local development, allowing you to run flows without needing lengthy CLI context flags.

### [Data Validation (Fail Fast)](./fail_fast)

**Concepts:** `required_context`, `Fail Fast`
Shows how to enforce data contracts using `required_context`. If a required key is missing, the engine stops immediately, preventing "silent failures" later in the flow.

### [Tools Demo (Safety & Metadata)](./tools-demo)

**Concepts:** `on_error`, `metadata.confirm_msg`, `Implicit IDs`
Demonstrates robust tool usage, including Safety Middleware (confirmation prompts) and Error Handling (`on_error` transitions).

---

## 3. Advanced Control

### [Global Signals (Interrupts)](./interrupt-demo)

**Concepts:** `on_signal`, `Graceful Shutdown`, `Ctrl+C`
Demonstrates how to handle global interruptions (like `SIGINT`) gracefully. Instead of crashing, the flow transitions to a confirmation node ("Are you sure?"), preventing data loss.

### [Context Injection (Testing & Automation)](./context-demo)

**Concepts:** `--context`, `Seed State`, `Templates`
Demonstrates how to inject initial data into the flow via the CLI flag `--context`. Critical for automated testing or integration with legacy systems.

---

## 4. Production & Observability

### [Structured Logging (Production)](./structured-logging)

**Concepts:** `slog` (JSON Logs), `Prometheus` (Metrics)
Demonstrates industry-standard observability by integrating Trellis with Go's `log/slog` and `prometheus/client_golang`. Shows how to output machine-readable logs and metrics.

### [Observability (Hooks & Debug)](./observability)

**Concepts:** `LifecycleHooks`, `--debug`, `Events`
Demonstrates how to use the `--debug` flag to visualize state transitions and events in the console.

---

## 5. Internals

### [Low Level API (Advanced)](./low-level-api)

**Concepts:** `Manual Loop`, `engine.Render`, `engine.Navigate`
Demonstrates how to manually drive the engine without using the `trellis.Runner`. Useful if you need to integrate Trellis into a custom UI framework, a game engine, or a non-standard event loop.
