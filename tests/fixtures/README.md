# Test Fixtures

This directory contains source code for auxiliary programs used during integration testing.

## Purpose

Trellis manages external processes as "Tools". To verify that the engine correctly handles various process states (success, failure, timeout, signal handling), we need real executables that behave in specific, predictable ways.

Instead of relying on system commands (like `sleep` or `false`) which vary across OS (Windows vs Linux), we compile these Go programs on-the-fly during the test run.

## Structure

- `resilience/`: Programs designed to test the robustness of the `process` adapter.

## Usage

These fixtures are typically compiled by helper functions in the test suite (e.g., `buildFixture` in `pkg/adapters/process/resilience_test.go`).
