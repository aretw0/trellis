# Architecture Proposal: Tool Result Persistence and Flattening

**Status**: **Accepted**
**Date**: 2026-03-06

## Context
Trellis needs a deterministic and ergonomic way to access the results of side-effects (tools) in subsequent nodes, especially for template interpolation.

Previously, users had to manually use `save_to: "key"` to persist the raw tool output. This was inconsistent and required manual knowledge of the return structure.

## Decision

1. **Engine-Managed Key**: The engine will automatically inject the most recent successful tool result into the context under the key `tool_result`.
2. **Flattening Logic**:
    * If the tool result is a `map[string]any`, its fields are merged directly into the `tool_result` map. The tool ID is stored as `_id`.
    * If the tool result is a scalar, it is stored under the `result` key (i.e., `{{ .tool_result.result }}`). The tool ID is still `_id`.
3. **Precedence (The "Collision" Policy)**:
    * Manual `save_to` logic in `applyInput` occurs *after* the engine-managed flattening in `handleToolResult`.
    * If a user explicitly sets `save_to: "tool_result"`, they are choosing to overwrite the engine-curated flattened map with the **raw** tool output.
    * **Recommendation**: Users should avoid using `save_to: "tool_result"` and instead use a different key if they want to persist the raw result specifically, or rely on the automatic `tool_result` for ergonomic access.

## Consequences

* Ergonomics: Templates can easily use `{{ .tool_result.some_field }}`.
* Transparency: The `_id` field is always available to identify which tool produced the result.
* Flexibility: Users can still shadow or override this behavior if they have advanced needs, although it is discouraged for standard usage.
