# Contributing to Trellis

Thank you for your interest in contributing to Trellis!

Trellis is an open-source project designed to be the robust backbone for AI Agents and Automation. We value clarity, reliability, and thoughtful design.

## Getting Started

1. **Clone the repository**:

    ```bash
    git clone https://github.com/aretw0/trellis.git
    cd trellis
    ```

2. **Sync dependencies**:

    ```bash
    go mod tidy
    ```

3. **Run tests**:

    ```bash
    make test
    # or
    go test ./...
    ```

## Development Standards

Before submitting a Pull Request, please review our standards:

### 1. Language Policy (Bilingual)

We follow a context-aware language policy:

- **English**: For all code (comments, variables), public documentation (`README`, `examples`), and user-facing CLI messages.
- **Portuguese**: For internal engineering documentation (`docs/TECHNICAL.md`, `docs/PLANNING.md`).

See [docs/STANDARDS.md](docs/STANDARDS.md) for details.

### 2. Testing

We have a high bar for reliability. All new features must be tested.
See [docs/TESTING.md](docs/TESTING.md) to understand our testing strategy and how to use Fixtures.

### 3. Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/).

- `feat: ...` for new features
- `fix: ...` for bug fixes
- `docs: ...` for documentation
- `refactor: ...` for code cleanup without behavior change

## Pull Request Process

1. Create a feature branch from `main`.
2. Ensure `make test` passes locally.
3. Open a PR describing *what* changed and *why*.
4. If your PR changes behavior, ensure `docs/` are updated.

## Community

Questions? Discussions? Feel free to open an Issue or start a Discussion on GitHub.
