# Trellis: Ecosystem Status

> **Papel**: State machine platform (6 layers) for AI agents & automation  
> **Última Atualização**: 2026-03-02

---

## Stack Overview

```
┌─────────────────────────────────────────────────────────┐
│ Layer 5: UI (TUI, Chat Web UI)                          │
├─────────────────────────────────────────────────────────┤
│ Layer 4: Protocols (HTTP, MCP, SSE)                     │
├─────────────────────────────────────────────────────────┤
│ Layer 3: Tooling (CLI, Registry, Watch, Introspection)  │
├─────────────────────────────────────────────────────────┤
│ Layer 2: DSLs (Flow-DSL: text, question, tool, prompt)  │
├─────────────────────────────────────────────────────────┤
│ Layer 1: Persistence (StateStore, Session, SAGA)        │
├─────────────────────────────────────────────────────────┤
│ Layer 0: Engine (Node, Scheduler, State, Templates)     │
└─────────────────────────────────────────────────────────┘
                    Built on
                      ↓
            ┌─────────────────────┐
            │ lifecycle (v1.5+)   │
            │ loam (v0.10+)       │
            │ introspection       │
            └─────────────────────┘
```

---

## Dependencies

---

## Dependencies

| Projeto | Versão | Status | Integração |
|---------|--------|--------|------------|
| **lifecycle** | v1.5+ | ✅ Mature | Signal handling (SIGINT/SIGTERM), Terminal I/O (Windows CONIN$), graceful shutdown |
| **loam** | v0.10+ | ✅ Mature | YAML/JSON/Markdown parser, watch mode, type-safe frontmatter |
| **introspection** | v0.1+ | ⚠️ Partial | Mermaid diagrams (internal generator), future: unified observability |
| **procio** | v0.1+ | ✅ Transitive | Process hygiene (PDeathSig, Job Objects) via lifecycle |

---

## Consumers

### 🌳 Arbour (Community Hub)

**Papel**: Package manager & registry for Trellis-based flows  
**Status**: Phase 1 ready (Package Manager MVP)  
**Integração**: Consumes Trellis as execution engine

- Flow execution via Trellis Runner
- Protocol adapters (HTTP, MCP, SSE)
- Lifecycle signal handling (inherited)

**Next**: Post Phase 2c — validate shared components (protocols, persistence, tooling)

---

### 🔮 Life-DSL (Future)

**Papel**: "Life as Code" DSL (habits, routines, energy management)  
**Status**: Planned (Phase 2b POC)  
**Objetivo**: Validate engine genericness before extraction

**Expected Integration**:

- ✅ Reuse: protocols, persistence, tooling (Layers 1, 3, 4)
- ❌ No reuse: flow-dsl node types (Layer 2)

---

## Roadmap

### ✅ Phase 1: Foundation Stabilization (2024-2025)

**Status**: Complete  
**Focus**: lifecycle v1.5, loam v0.10, introspection v0.1  
**Result**: Trellis v0.7+ stable, in production (Arbour)

---

### 🔧 Phase 2a: Internal Restructuring (2-3 weeks) [NEXT]

**Status**: Planned  
**Focus**: Organize Trellis in layers (`engine/`, `flow/`, `protocols/`) without repo extraction  
**Goal**: Clean code, 100% backward compat, ground zero for validation

**Blocker**: Node abstraction design decision

---

### 🧪 Phase 2b: Life-DSL POC (2-3 weeks)

**Status**: Planned  
**Focus**: Implement life-dsl **inside Trellis** to validate genericness  
**Goal**: Empirical discovery — what's truly generic?

**Blocker**: Phase 2a completion

---

### 📦 Phase 2c: Surgical Extraction (1-2 weeks)

**Status**: Planned  
**Focus**: Extract **only** validated shared components  
**Goal**: Separate repos with **proven** reusability

**Blocker**: Phase 2b validation

---

### 🚀 Phase 3: Life-DSL Standalone (2-3 weeks)

**Status**: Planned  
**Focus**: Separate repo `aretw0/life-dsl`  
**Goal**: Validate Abstract Engine architecture with 2 DSLs

---

### 🎯 Phase 4: Extract Trellis-Engine (2027+)

**Status**: Awaiting validation  
**Focus**: Extract 100% generic core engine after 2+ DSLs in production  
**Philosophy**: "Measure twice, cut once"

---

**Next Review**: After Phase 2a (internal refactor complete)
