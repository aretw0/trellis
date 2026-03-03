# Trellis Refactoring: Key Insights Handoff

> **Source**: Transferred from lifecycle ecosystem analysis (2026-03-02)  
> **Context**: Pre-refactoring analysis to preserve valuable insights during engine abstraction.

## Executive Summary

Trellis v0.7+ is **not just a state machine engine** — it's a complete platform with 5 architectural layers. Before extracting the abstract engine, preserve these 10 critical insights.

---

## 10 Preserved Insights

### 1. Hypertext State Machine

**Insight**: Markdown links become implicit transitions.

```markdown
# Welcome
[Log In](./login.md) or [Register](./register.md)?
```

**Impact**: Drastically lowers barrier to entry.  
**Scope**: Generic — any DSL can use this pattern.

### 2. File-System Routing (Web Server)

**Insight**: `repo/about.md` → `/about` node (Astro/Next.js style).  
**Impact**: Trellis becomes a "Stateful Web Server" with HATEOAS.  
**Scope**: Generic protocol adapter — reusable across DSLs.

### 3. MCP Server (AI-First)

**Insight**: LLMs orchestrate flows via Model Context Protocol.  
**Impact**: AI agents can navigate/render/graph flows.  
**Scope**: Generic adapter — life-dsl, scrape-dsl can reuse.

### 4. Chat Web UI with SSE Delta Patching

**Insight**: Zero-config reactive UI at `/ui`.  
**Impact**: Instant visualization for any flow.  
**Scope**: Partially generic (reactivity = yes, chat theme = flow-specific).

### 5. Polyglot Tool Registry

**Insight**: Execute `.sh`, `.py`, `.js`, `.ps1` via Unix contract (ENV, STDIN, STDOUT).  
**Impact**: Trellis as "Polyglot Orchestrator".  
**Scope**: 100% generic — extract to `trellis-tooling`.

### 6. Session Management & SAGA Patterns

**Insight**: Durable execution via snapshotting (no Event Sourcing).  
**Impact**: Long-running flows survive restarts.  
**Scope**: Generic — extract to `trellis-persistence`.

### 7. Schema Validation (`context_schema`)

**Insight**: Type-check variables before runtime.  
**Impact**: Catch errors early.  
**Scope**: Generic — part of abstract engine.

### 8. Type-Safe Builders (Go DSL)

**Insight**: Fluent API avoids YAML verbosity.  
**Scope**: Pattern reusable — each DSL implements its own builder.

### 9. Signal Differentiation

**Insight**: SIGINT ≠ SIGTERM (User Interrupt vs System Termination).  
**Status**: ✅ Already in `lifecycle` v1.5.  
**Action**: Upgrade Trellis from lifecycle v0.1.1 → v1.7+.

### 10. Entrypoint Fallback (`start` → `main` → `index`)

**Insight**: Support multiple conventions.  
**Scope**: DSL-specific (configurable, not hardcoded).

---

## Refactoring Strategy (Phases 2a/2b/2c)

### Phase 2a: Internal Restructuring (2-3 weeks)

**Goal**: Organize Trellis into clear layers **without extracting repos**.

```text
trellis/
├── pkg/
│   ├── engine/       ← Core execution (generic)
│   ├── flow/         ← Flow-DSL specifics
│   ├── protocols/    ← HTTP, MCP, SSE (generic)
│   ├── persistence/  ← State store, sessions (generic)
│   ├── tooling/      ← Tool registry (generic)
│   └── ui/           ← Patterns (generic) + themes (specific)
└── trellis.go        ← Backward compat facade
```

**Validation**: All tests pass, Arbour continues working.

### Phase 2b: Life-DSL POC (2-3 weeks)

**Goal**: Implement `life-dsl` **inside trellis repo** to discover what's truly generic.

```text
trellis/pkg/
├── engine/  ← Shared by flow + life
├── flow/    ← Flow-DSL specifics
└── life/    ← Life-DSL experiment
    ├── types.go      (workers, habits)
    ├── compiler.go   (life.yaml → engine)
    └── executors.go  (CLI, Browser, Notify)
```

**Questions to Answer**:

- Does life-dsl need HTTP server? → protocols/ is generic ✓
- Does life-dsl need sessions? → persistence/ is generic ✓
- Does life-dsl need tool registry? → tooling/ is generic ✓
- Different node types? → engine/ needs more abstraction

### Phase 2c: Surgical Extraction (1-2 weeks)

**Goal**: Extract **only** validated generic components.

**Extract**:

- `trellis-protocols` (HTTP, MCP, SSE)
- `trellis-persistence` (StateStore, Session)
- `trellis-tooling` (Tool Registry, Process Adapter)

**Keep Monolithic**:

- `pkg/engine/` (needs more iteration)
- `pkg/ui/` (themes are flow-specific)
- `pkg/dsl/` (each DSL has its own)

---

## Critical Design Decisions

### Node Abstraction: Hybrid Approach Recommended

```go
// Flexibility (functions) + Safety (interface) + DX (builder)
type Node struct {
    id       string
    execute  func(ctx context.Context) error
    schedule Scheduler
    onStart  func(ctx context.Context) error
    onFailure func(ctx context.Context, err error) error
}

type Executable interface {
    ID() string
    Execute(ctx context.Context) error
    Scheduler() Scheduler
}

// Builder pattern for ergonomics
func NewNode(id string) *NodeBuilder {
    return &NodeBuilder{node: &Node{id: id}}
}
```

**Why Hybrid?**

- Functions = composable, testable
- Interface = engine contract guaranteed
- Builder = idiomatic Go, readable
- Extensible = add hooks without breaking

### Scheduler Interface

```go
type Scheduler interface {
    Next(ctx context.Context, current Node) (Node, error)
}

// Implementations:
// - CronScheduler (time-based for life-dsl)
// - StateMachineScheduler (transition-based for flow-dsl)
// - SelectorScheduler (DOM traversal for scrape-dsl)
```

### State Store Interface

```go
type StateStore interface {
    Get(ctx context.Context, nodeID string) (*NodeState, error)
    Set(ctx context.Context, nodeID string, state *NodeState) error
    Checkpoint(ctx context.Context) error
    Restore(ctx context.Context) error
}

// Implementations: MemoryStore, SQLiteStore, RedisStore, LoamStore
```

---

## Integration with Lifecycle (Post-Refactor)

**Current**: Trellis uses `lifecycle v0.1.1` (basic cancellation only).

**Target**: Deep integration with lifecycle v1.5+ (Control Plane).

```go
type Engine struct {
    router     *lifecycle.Router      // Event routing
    supervisor *lifecycle.Supervisor  // Worker orchestration
    store      StateStore             // Persistence
}

func (e *Engine) Run(ctx context.Context, nodes []Executable) error {
    ctx = lifecycle.Attach(ctx, e.router)
    
    for _, node := range nodes {
        worker := e.supervisor.Add(node.ID(), node.Execute)
        worker.RestartPolicy = lifecycle.Always
    }
    
    return lifecycle.Run(ctx, e.supervisor)
}
```

**Benefits**:

- Suspend/Resume flows
- Graceful shutdown with checkpointing
- Signal differentiation (SIGINT vs SIGTERM)
- Event-driven control (webhooks, file watches, health checks)

---

## Action Items for Trellis Maintainers

### Before Phase 2a

- [ ] Review this document and consolidate with PLANNING.md
- [ ] Create `trellis/docs/ECOSYSTEM_INTEGRATION.md` (use lifecycle's as template)
- [ ] Audit current architecture vs 5-layer stack
- [ ] Identify all coupling points (DSL ↔ Engine)

### During Phase 2a

- [ ] Restructure `pkg/` into engine/flow/protocols/persistence/tooling/ui
- [ ] Maintain 100% backward compatibility
- [ ] Run full test suite after each move
- [ ] Update Arbour integration tests

### During Phase 2b

- [ ] Implement life-dsl experiment in `pkg/life/`
- [ ] Document friction points (what's hard to reuse?)
- [ ] Validate generic vs specific boundaries
- [ ] Create concrete extraction checklist

### During Phase 2c

- [ ] Extract validated components to separate repos
- [ ] Update import paths with backward compat aliases
- [ ] Publish extraction ADR
- [ ] Update ecosystem documentation

---

## References

- [trellis/PLANNING.md](./PLANNING.md) — Project roadmap
- [trellis/docs/ECOSYSTEM_INTEGRATION.md](./ECOSYSTEM_INTEGRATION.md) — Integration with lifecycle ecosystem
- [lifecycle/docs/ecosystem/engine_abstraction.md](https://github.com/aretw0/lifecycle/blob/main/docs/ecosystem/engine_abstraction.md) — Full vision & design

---

**Last Updated**: 2026-03-02  
**Status**: Ready for Phase 2a implementation  
**Next Step**: Trellis maintainer to decide on Node abstraction (see Section: Critical Design Decisions)
