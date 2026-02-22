# Architecture Proposal: Trellis Single-File Execution & Automation

**Status**: **Accepted**
**Date**: 2026-02-22

* **Context**: The `go-rod/wayang` project demonstrated a controller over a DevTools driver (Rod) using JSON/declarative actions to perform web scraping and automation. Historically, `trellis` has focused on complex, multi-file neuro-symbolic flows (directories with `start.md`, schemas, scripts). However, to fully serve as a robust automation engine and a viable successor to stagnant projects like `wayang`, `trellis` must support minimal, low-friction entry points.
* **Decision**:
    1. **Single-File Loader**: The Trellis CLI (`trellis run`) must support executing a single `.yaml`, `.json`, or `.md` file directly, without requiring a full directory structure.
    2. **Web Automation Domain**: Trellis should formalize standard "Action Nodes" (or an adapter plugin) inspired by Rod/Wayang (e.g., `navigate`, `click`, `extract` with automatic waits and chained contexts) to natively support end-to-end web scraping scenarios without requiring users to write custom Go code for simple bots.
* **Rationale**:
    1. **Developer Experience**: Scripts for web scraping are often single-file. Forcing a directory structure creates unnecessary cognitive load for simple automations.
    2. **Ecosystem Domination**: By lowering the barrier to entry for single-file scripts and providing robust web action bindings, Trellis can absorb the use cases of `wayang` and generic RPA (Robotic Process Automation) tools.
* **Consequences**:
  * The `loam` loader (or the Trellis CLI wrapper) needs an update to detect if the target path is a file or a directory.
  * Creation of a `trellis-rod` plugin or specific Node Types for WebDriver/CDP actions.
