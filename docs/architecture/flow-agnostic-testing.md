# Architecture Proposal: Flow-Agnostic Testing via State Space Exploration (Fuzzing)

**Status**: **Proposed** / **Vision**
**Date**: 2026-02-22

* **Context**: Currently, verifying that a Trellis flow works as expected relies on manual execution or writing brittle integration tests that simulate HTTP/CLI inputs exactly matching the flow's expected answers. As the visual "Chat UI/Inspector" becomes a core part of the engine, there is a risk of coupling UI integration tests to specific, fragile flows (like `examples/tour`). Furthermore, users building complex flows currently have no automated way to guarantee that their flow won't crash in an obscure edge-case node without writing exhaustive manual tests covering every branch.
* **Decision**:
    1. **Exhaustive Feature Fixtures (Short Term)**: For testing the embedded Chat UI, we will NOT use real user flows. Instead, we will create a synthetic, minimal flow (`tests/fixtures/ui_exhaustive`) explicitly designed to emit every possible action type (Text, Input Prompt, Tool Call, Error, End) in a predictable sequence. The UI tests (via `go-rod`) will navigate this fixture to verify the *UI Contract*, remaining completely decoupled from business logic.
    2. **Fuzzing Engine (Long Term Vision)**: Trellis will introduce a `trellis test` or `trellis fuzz` capability. Since Trellis uses a Deterministic Finite Automaton (DFA), the Engine will automatically explore the entire state space of a flow. When it encounters nodes requiring input, the testing harness will automatically generate mock data satisfying the required schemas.
* **Rationale**:
    1. **Decoupling**: Testing UI rendering against synthetic fixtures ensures UI integration tests are lightning fast and isolated from changes to real-world examples.
    2. **Behavioral Guarantees**: Auto-fuzzing will guarantee that a flow can run from start to finish without panics, infinite loops, or invalid internal state mutations, entirely agnostic of the actual business logic or UI rendering.
* **Consequences**:
  * **Positive**: We have a documented roadmap for providing massive value to users via automated graph testing.
  * **Negative**: Creating synthetic fixtures requires maintaining them alongside new Trellis engine features (e.g., if a new node type is added, the exhaustive fixture must be updated).
