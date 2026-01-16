# GEMINI.md

## Project Overview

Trellis is a deterministic state machine engine written in Go, designed for building CLIs, automation, and AI agent guardrails.

It follows a **hexagonal architecture (ports and adapters)**, decoupling the core logic from implementation details like file I/O. The `pkg/` directory contains the core domain and interfaces, while `internal/` holds concrete implementations. This allows Trellis to be used as a library or as a standalone CLI.

The default implementation uses **Loam**, a library that reads Markdown/JSON/YAML files to define the state machine graph.

## Key Commands

To get started, ensure Go dependencies are synced:

```bash
go mod tidy
```

### Running the Demo

The project includes a guided tour. To run it:

```bash
go run ./cmd/trellis run ./examples/tour
```

### Common Usage

- **Run a flow interactively:**

    ```bash
    go run ./cmd/trellis run <path_to_flow_directory>
    ```

- **Run in headless mode for automation:**

    ```bash
    echo "\"input1\"\n\"input2\"" | go run ./cmd/trellis run --headless <path_to_flow_directory>
    ```

- **Visualize a flow (outputs Mermaid graph):**

    ```bash
    go run ./cmd/trellis graph <path_to_flow_directory>
    ```

- **Develop with hot-reload:**

    ```bash
    go run ./cmd/trellis run --watch --dir <path_to_flow_directory>
    ```

## Development Philosophy

- **Decoupled Core:** The engine's core logic (`pkg/`) is pure and has no knowledge of the outside world. Adapters in `internal/` connect it to external systems (like the Loam file loader). When modifying, respect this separation.
- **CLI as Primary Interface:** The `trellis` CLI is the main entry point for most tasks. Enhancements should be exposed as CLI commands where appropriate.
- **Convention over Configuration:** The engine relies on file naming conventions (e.g., `start.md`) and simple formats. Adhere to these conventions when creating new flows.
- **Test-First for Adapters:** When modifying adapters (like `LoamLoader`), always modify or add the corresponding contract test first to ensure behavior consistency across all implementations.
