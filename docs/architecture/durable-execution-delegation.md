# ADR: Delegation of Durable Execution to Lifecycle

## 1. Context

In Trellis `v0.6` (The "Durable" Phase), we introduced state persistence to support long-running workflows, serverless execution, and paused states. This required building orchestration primitives internally:

* `pkg/session`: A Session Manager responsible for Two-Level Locking (Local `sync.Mutex` + Distributed Redis Locks) and concurrency control (Reference Counting).
* `StateStore`: An interface for persistence adapters (File, Redis).

However, managing distributed locks, polling for wake-ups, and handling robust event streams (like timeouts, cron schedules, or asynchronous webhooks that survive reboots) significantly increases the cognitive load of a library designed to be a "Deterministic Finite Automaton (DFA)".

Meanwhile, our sister project **[lifecycle](https://github.com/aretw0/lifecycle)** is explicitly evolving into a Control Plane for infrastructure (ADR-0014: Durable Extension for the Event Router, ADR-0015: Worker Role Grouping).

## 2. Decision

**Proposal**: Demote Trellis from a "Distributed Orchestrator" back to a pure, stateless "Neuro-Symbolic Engine", and delegate all orchestration, durable event routing, and distributed state management to `lifecycle`.

Instead of Trellis pulling events or managing locks:

1. `lifecycle` acts as the Durable Event Broker and Supervisor.
2. `trellis` acts as a pure `lifecycle.Worker` plugin.
3. When an event (e.g., a webhook or a timer) arrives, `lifecycle` wakes up the appropriate Trellis session, injects the event into the Trellis Engine (`Navigate`), and persists the resulting state.

## 3. Rationale (The Pivot)

By adopting this separation of concerns:

* **Trellis Remains Pure**: Trellis codebase shrinks, dropping complex synchronization (Mutexes, Redis ZSET polling, Distributed Locks) and remaining purely focused on Graph Traversal, Validation, LLM Interop (MCP), and SAGA tracking.
* **Lifecycle Gets a Killer Consumer**: Features like "Worker Role Grouping" in `lifecycle` directly serve Trellis's need to route execution of a specific session to a specific pod in a cluster.
* **Ecosystem Synergy**: The `trellis` tool becomes a lightweight execution container for `lifecycle` event loops.

## 4. Consequences

1. **Deprecation Path**: `pkg/session` (and its Two-Level Locking) may be deprecated or heavily refactored to just be an in-memory session cache for CLIs.
2. **Persistence Adapters**: `pkg/adapters/redis` and `pkg/adapters/file_store` might move out of Trellis and into the `lifecycle` or an overarching "Orchestrator" service.
3. **Execution Model**: The `Runner` becomes strictly single-shot or acts purely as a `lifecycle.Worker` receiving events via Go channels natively (`events.Notify(ch)`).
4. **Adoption**: This positions the ecosystem perfectly for Cloud-Native deployments (Kubernetes) where `lifecycle` handles the Pod mechanics, and `trellis` handles the business logic.
