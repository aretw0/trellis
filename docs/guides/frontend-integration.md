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
```

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
