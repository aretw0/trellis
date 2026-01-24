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

## 3. Infrastructure & Patterns

### [Manual Security (Encryption & PII)](./manual-security)

**Concepts:** `Middleware`, `Encryption`, `PII Masking`
Showcases how to securely wrap the state store with encryption and PII sanitization middleware. This is a manual infrastructure setup (wrapping the runner) rather than a built-in feature.

### [Compensation Manual (Transaction Rollback)](./compensation-manual)

**Concepts:** `Saga Pattern`, `on_error`, `Rollback`
Demonstrates how to implement distributed transactions with compensating actions (Rollback) using standard Trellis primitives (Manual Wiring).

### [Compensation Native (Automatic Rollback)](./compensation-native)

**Concepts:** `Native SAGA`, `do/undo`, `on_error: rollback`
Demonstrates the **new** Native SAGA orchestration in Trellis v0.7. The engine automatically handles the stack unwinding and compensation execution when a failure occurs.

### [Compensation Durable (Long Running)](./compensation-durable)

**Concepts:** `Durable Execution`, `Persistence`, `Signals`
Demonstrates a **Long Running SAGA**. The flow pauses for an external signal (Manager Approval), is interrupted (process exit), resumed from disk, and then rolled back upon Rejection.

### [Confirm Demo (Unix Style)](./confirm-demo)

**Concepts:** `input_type: confirm`, `input_default`, `Unix Conventions`
Demonstrates the native confirmation UX. It shows how the engine follows standard CLI conventions (Enter = Yes) and how to formally override those defaults for secure flows.

### [Process Demo (Dynamic Scripts)](./process-demo)

**Concepts:** `x-exec`, `Tools`, `YAML Metadata`
Demonstrates the power of Universal Action Semantics. This example shows how to define and execute dynamic OS processes (scripts) directly from node metadata without pre-compiling Go tools.

---

## 4. Advanced Control

### [Global Signals (Interrupts)](./interrupt-demo)

**Concepts:** `on_signal`, `Graceful Shutdown`, `Ctrl+C`
Demonstrates how to handle global interruptions (like `SIGINT`) gracefully. Instead of crashing, the flow transitions to a confirmation node ("Are you sure?"), preventing data loss.

### [Sub-graph Demo (Modularity)](./subgraph-demo)

**Concepts:** `Modular Graphs`, `Portability`, `Logical Segregation`
Demonstrates how to split a complex state machine into multiple smaller, manageable files. This is the cornerstone of building complex, enterprise-ready automation.

### [Reload Demo (Live Dev)](./reload-demo)

**Concepts:** `--watch`, `Hot Reload`, `DX`
Shows how Trellis supports a high-velocity developer experience. Modify your Markdown files and see the engine reload the logic in real-time without restarting the process.

### [Context Injection (Testing & Automation)](./context-demo)

**Concepts:** `--context`, `Seed State`, `Templates`
Demonstrates how to inject initial data into the flow via the CLI flag `--context`. Critical for automated testing or integration with legacy systems.

---

## 5. Production & Observability

### [Signals Demo (Interrupts & Timeouts)](./signals-demo)

**Concepts:** `on_signal`, `step_timeout`
Demonstrates global signals (interrupts) and step timeouts.

### [Structured Logging (Production)](./structured-logging)

**Concepts:** `slog` (JSON Logs), `Prometheus` (Metrics)
Demonstrates industry-standard observability by integrating Trellis with Go's `log/slog` and `prometheus/client_golang`. Shows how to output machine-readable logs and metrics.

### [Observability (Hooks & Debug)](./observability)

**Concepts:** `LifecycleHooks`, `--debug`, `Events`
Demonstrates how to use the `--debug` flag to visualize state transitions and events in the console.

---

## 6. Internals & API

### [Low Level API (Advanced)](./low-level-api)

**Concepts:** `Manual Loop`, `engine.Render`, `engine.Navigate`
Demonstrates how to manually drive the engine without using the `trellis.Runner`. Useful if you need to integrate Trellis into a custom UI framework, a game engine, or a non-standard event loop.
