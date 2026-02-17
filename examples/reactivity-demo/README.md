# Reactivity Demo

This example demonstrates the **Real-Time Reactivity** features introduced in Trellis v0.7.9. It uses Server-Sent Events (SSE) to push granular state updates (Deltas) to the frontend, eliminating the need for polling.

## Files

- `start.md`: The entry point (defines transition to finish).
- `finish.md`: The exit point.
- `index.html`: A vanilla JS frontend that consumes the SSE stream.

## How to Run

### 1. Start the Server

Run the Trellis HTTP server pointing to this directory:

```bash
# From project root
go run ./cmd/trellis serve --dir ./examples/reactivity-demo --port 8080
```

### 2. Open the Frontend

Open `index.html` in your browser. You can do this by dragging the file into Chrome/Firefox or keeping it simple:

> **Note**: Modern browsers may block SSE connections from `file://` protocol due to CORS/Security policies. It is recommended to serve the HTML file via a simple HTTP server or use the "Open File" approach if your browser permits.

If needed, serve the directory:

```bash
npx serve ./examples/reactivity-demo
# Open http://localhost:3000
```

### 3. Usage

1. **Connect**: The page automatically generates a Session ID and connects to `http://localhost:8080/events`.
2. **Observe**: Watch the "Event Stream" panel. You should see a `ping` event.
3. **Interact**: Click **Send 'next'**.
4. **React**:
    - The server processes the input.
    - The server calculates the difference (Diff).
    - The server pushes a JSON Delta to the client.
    - The client patches its local state and updates the UI.

## Key Concepts

- **State Diffing**: The server only sends what changed (e.g., `{"context": {"foo": "bar"}}`), not the whole state.
- **Client Patching**: The frontend merges these updates into its local replica.
