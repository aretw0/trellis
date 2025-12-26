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
