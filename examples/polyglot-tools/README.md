# Polyglot Tools Example (Trellis Process Adapter)

This example demonstrates how to integrate scripts written in **Python**, **Node.js**, and **PowerShell** with Trellis.

## The Contract (Unix Style)

Trellis adheres to a simple, universal contract for external tools, inspired by the Unix philosophy:

1. **Input**: Arguments are passed as **Environment Variables**, prefixed with `TRELLIS_ARG_`.
    - Example: `args: { name: "Bob" }` becomes `TRELLIS_ARG_NAME="Bob"`.
    - This avoids CLI flag parsing complexity and injection risks.
2. **Output**: The tool should print a **JSON Object** to `Stdout`.
    - Trellis automatically detects JSON output and parses it into a structural object.
3. **Error**: To signal failure, exit with a **non-zero status code**.
    - Optional: Print error details to `Stderr`, which Trellis captures for debugging.

## Directory Structure

- `tools.yaml`: Registers the scripts as named tools (`greet_py`, `greet_js`, `greet_ps`).
- `py-tool.py`: Python implementation (using `os.environ` and `print(json.dumps)`).
- `js-tool.js`: Node.js implementation (using `process.env` and `console.log`).
- `ps-tool.ps1`: PowerShell implementation (using `$env:` and `ConvertTo-Json`).
- `start.md`: The entry point flow that orchestrates the execution.

## How to Run

Navigate to this directory and run Trellis:

```bash
# Run interactively
go run ../../cmd/trellis run .

# Run headless (for testing)
go run ../../cmd/trellis run . --headless
```

## Tips for Windows Users

- Ensure `python`, `node`, and `pwsh` (or `powershell`) are in your PATH.
- If using `tools.yaml`, the `command` field must match an executable on your system.
