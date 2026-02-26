# Testing Trellis

This document outlines the testing strategy, structure, and guidelines for the Trellis project.

## Running Tests

Trellis uses the standard Go testing toolchain.

### Quick Start

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

### Using Make

We provide a `Makefile` for common tasks:

- `make test`: Runs all unit and integration tests.
- `make coverage`: Runs tests and generates a coverage report.

## Test Structure

The codebase is committed to a high standard of reliability, employing a mix of unit, integration, and resilience tests.

### 1. Unit Tests (`*_test.go`)
Located alongside the source code they test (e.g., `pkg/domain/state_test.go`). These test individual functions and structs in isolation.

### 2. Integration Tests (`tests/`)
The `tests/` directory contains black-box integration tests that verify the system as a whole. These tests typically:

- Initialize a full `runner.Runner` or `trellis.Engine`.
- Load real or synthetic graphs.
- Execute full lifecycle scenarios.

### 3. Stability & Resilience (`tests/stability`, `pkg/adapters/process`)
We heavily test for stability under adverse conditions.

- **Resilience:** Verifies that the engine handles crashes, timeouts, and signals correctly.
- **Stability:** Long-running tests (fuzzing-like) to detect memory leaks or race conditions.

## Test Fixtures (`tests/fixtures`)

To test process management and resilience, we use **Fixtures**.

These are small, collaborative Go programs located in `tests/fixtures`.
During testing, these programs are compiled on-the-fly and executed by the Trellis Process Adapter.

They simulate:

- **Good Citizens:** Processes that handle signals and exit gracefully.
- **Bad Citizens:** Processes that ignore signals, hang, or crash.

## Cross-Platform Quirks

Trellis is designed to be cross-platform, but some adapters (especially `process` and `ui`) have OS-specific behaviors that are documented here for troubleshooting.

### 1. Process Signal Propagation (Windows vs Unix)
The `process` adapter uses `os.Interrupt` (SIGINT) for graceful shutdowns.

- **Unix/Linux/WSL**: Signals propagate correctly down the process tree.
- **Windows**: `os.Interrupt` is only sent to the console process group. Background processes started by the `process` adapter may not receive the signal correctly and will often fall back to the **Force-Kill** grace period (default 5s).

### 2. UI Tests & CI Sandboxing
UI tests use `go-rod` to control a headless Chromium instance.

- **GitHub Actions (Linux)**: Most CI runners do not allow the default Chromium sandbox. Tests automatically detect the `GITHUB_ACTIONS=true` environment variable and apply the `--no-sandbox` flag only when strictly necessary.
- **Local Dev**: In local environments (Windows/macOS/Linux), the browser runs with its default sandbox enabled for maximum security and fidelity.
- **Windows AppData**: We disable `Leakless` in tests to avoid extraction issues in restricted environments.

### 3. Data Race Mitigation
Capturing output from rapidly spawning and terminating processes can trigger data races if buffers are not synchronized. The `process` adapter uses a thread-safe `safeBuffer` to protect `stdout` and `stderr` during reads in the `Execute` method.

## Philosophy

- **Test Implementation, Not Mocking:** We prefer using real adapters (like `loam`, `process`) over mocks where feasible.
- **Behavior Drivien:** Tests should verify the *behavior* of the system (e.g., "The state transitioned to 'done'") rather than internal implementation details.
- **Deterministic:** Tests must be deterministic. Avoid `time.Sleep` where possible; use channels and synchronization primitives.
