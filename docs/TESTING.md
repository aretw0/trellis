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

See [tests/fixtures/README.md](tests/fixtures/README.md) for details.

## Philosophy

- **Test Implementation, Not Mocking:** We prefer using real adapters (like `loam`, `process`) over mocks where feasible.
- **Behavior Drivien:** Tests should verify the *behavior* of the system (e.g., "The state transitioned to 'done'") rather than internal implementation details.
- **Deterministic:** Tests must be deterministic. Avoid `time.Sleep` where possible; use channels and synchronization primitives.
