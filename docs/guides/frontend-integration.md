# Frontend Integration Guide

This guide explains how to build a reactive frontend that connects to the Trellis HTTP Server.

## 1. Connect to the Event Stream

Use the native `EventSource` API to subscribe to state updates. You MUST provide a unique `session_id`.

```javascript
const sessionId = "sess-" + Math.random().toString(36).substr(2, 9);
const evtSource = new EventSource(`/events?session_id=${sessionId}`);

evtSource.onmessage = (event) => {
    const delta = JSON.parse(event.data);
    applyDelta(delta);
};
```

## 2. Managing State (Client-Side Patching)

The server sends **Deltas** (changes), not full snapshots. You need to maintain a local state replica and patch it.

```javascript
let localState = {
    session_id: sessionId,
    context: {},
    history: []
};

function applyDelta(delta) {
    // 1. Update Scalar Fields
    if (delta.current_node_id) localState.current_node_id = delta.current_node_id;
    if (delta.status) localState.status = delta.status;

    // 2. Patch Context (Merge)
    if (delta.context) {
        localState.context = { ...localState.context, ...delta.context };
    }

    // 3. Patch History (Append)
    if (delta.history && delta.history.appended) {
        localState.history.push(...delta.history.appended);
    }
    
    renderUI();
}

## Interactive Elements

When the engine reaches a node requiring input (e.g., `input_type: text`), the response will include `state.status: "waiting_for_input"`. Use the `Prompt` schema to render appropriate UI elements (text fields, buttons, etc.).

---

## Internationalization (I18n)

Trellis supports first-class internationalization through the `messages` property on nodes.

### Serving Localized Content
The engine can return localized messages if provided in the node definition. It is recommended that clients send a `Accept-Language` header or a `locale` field in context to help the engine decide, though the logic is often client-side in simple implementations.

### Inspector I18n System
The Trellis Inspector UI uses a robust client-side I18n system:
- **Dictionary-based**: EN, PT-BR, and ES support.
- **Auto-detection**: Uses `navigator.language` with fallback.
- **Persistence**: Remembers user choice via `localStorage`.

---

## Accessibility (WCAG 2.1 AA)

The Trellis Inspector is designed with accessibility as a core requirement:

### ARIA Landmarks
- `<header role="banner">`: Navigation and branding.
- `<main role="main">`: Primary workspace.
- `<div role="log" aria-live="polite">`: Live chat history updates.

### Color Contrast
- All text meets **WCAG 2.1 AA** contrast ratios (e.g., Slate-600/700 on white).
- Interactive elements have distinct focus states (`focus:ring-2`).

### Keyboard Navigation
- Full keyboard support for input and signals.
- **Skip-to-content** link available for screen readers.

---

## Markdown & Content Conversion

Trellis allows plugging in a `ContentConverter` (e.g., for Markdown to HTML).

### Backend Configuration
```go
engine := runtime.NewEngine(
    runtime.WithContentConverter(adapters.NewMarkdownConverter()),
)
```

### Frontend Rendering
Nodes of `type: format` with `format: markdown` will have their content pre-rendered by the engine if a converter is present, or remain as raw Markdown for the frontend to handle.

## 3. Triggering Transitions

To change state, send a POST request to `/navigate`. You don't need to wait for the response to update UI if you trust the optimistic update, but typically you wait for the SSE event to be the "source of truth".

```javascript
async function sendInput(input) {
    await fetch("/navigate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            state: localState, // Send current state to server
            input: input
        })
    });
}
```

## Example

See `examples/reactivity-demo/index.html` for a complete, zero-build working example.
