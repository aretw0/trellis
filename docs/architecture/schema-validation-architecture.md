# Architecture Decision: Typed Flows - Schema Validation

**Status**: ✅ **DECIDED** (Option A with Extraction Path)  
**Date**: 2026-02-16  
**Affected Components**: `pkg/schema`, `pkg/adapters/loam`, `internal/runtime/engine`, `pkg/domain`  
**Related Task**: [PLANNING.md § v0.7.5](../PLANNING.md) - "Typed Flows"  
**Tracking**: [DECISIONS.md](../DECISIONS.md)

---

## Executive Summary

**Decision**: Implement **Option A (Validation in Trellis)** with **Extraction Path** for possible future lib separation.

**Rationale**:

1. ✅ **Zero Loam Coupling**: Schema is apenas um metadata field, Loam not affected
2. ✅ **Pragmatic**: No assumptions about ecosystem reuse now. Wait for evidence of traction.
3. ✅ **Door Open**: `pkg/schema` is standalone, trivial to extract as `github.com/aretw0/schema` v0.8+ if usage grows
4. ✅ **Incremental**: Start with basic types (string, int, float, bool, array), extend based on need

**Naming Strategy**:

- **v0.7.5**: Implement in `pkg/schema/` (flexible, generic naming)
- **v0.8+**: If traction evident (JSON Schema, coercion, code gen), propose extraction with clear scope
- **If no traction**: Stays internal, stable, no breaking commitment

---

## 1. Problem Statement

**Objetivo**: Permitir que grafos Trellis definam schemas estritos para o `Context` (ex: `api_key: string`, `retries: int`) e validem tipos no carregamento e runtime.

**Motivação**:

- **Type Safety**: Falhas no tipo de dados ficam óbvias e rápidas (fail-fast).
- **Documentation**: Schema actua como contrato explícito do fluxo.
- **Incremental Growth**: Fundação pra crescer (JSON Schema, coerção) se houver tração.

**Contexto Atual**:

- Trellis já suporta `required_context: ["api_key"]` (validação de presença).
- Loam é um "Embedded Transactional Engine" agnóstico, com `TypedRepository[T]` para metadata tipada.
- Trellis usa `TypedRepository[NodeMetadata]` para carregar nós.

---

## 2. Decision & Recommendation

### ✅ **Chosen: Option A (Validation in Trellis) with Extraction Path**

**Why Option A**:

- **Minimalism**: Zero Loam changes, zero new APIs, focused on Trellis use case.
- **Risk-Free**: If ecosystem reuse never materializes, no technical debt.
- **Flexible**: `pkg/schema` can stay internal forever or extract anytime.

**Why WITH Extraction Path**:

- **Future Option**: If usage grows (JSON Schema, coerção, generative tooling), extraction is ~1 day work.
- **No Overcommit**: Naming is internal (generic `pkg/schema`), no promise on pub lib name yet.
- **Learn & Decide**: Based on real Trellis usage patterns, make informed choice about scope expansion.

### ❌ Rejected Alternatives

**Option B (Loam Hooks)**: Adds scope creep to Loam for uncertain benefit. Validação em load-time não resolve runtime context mismatch.

**Option C (Separate Package Now)**: Premature; assumes reuse antes de ver tração no Trellis.

---

## 3. Implementation Plan (Immediate v0.7.5)

### Phase 1: Core Schema Package (~1 day)

Create `pkg/schema/` with standalone, zero-dependency types:

```
pkg/schema/
├── types.go        # Type interface + built-in implementations
├── validate.go     # Validate(schema, data) error
├── errors.go       # ValidationError struct
└── types_test.go   # Unit tests
```

**Key Files** (no Trellis internals):

```go
// pkg/schema/types.go
package schema

// Type defines the contract for field validation
type Type interface {
  Name() string
  Validate(value any) error
}

// Example built-in implementations
func String() Type { /* ... */ }
func Int() Type { /* ... */ }
func Float() Type { /* ... */ }
func Bool() Type { /* ... */ }
func Slice(elemType Type) Type { /* ... */ }
func Custom(name string, validate func(any) error) Type { /* ... */ }

// Schema is a map of field validators
type Schema map[string]Type

// Validate checks all fields
func Validate(schema Schema, data map[string]any) error {
  for key, t := range schema {
    value, ok := data[key]
    if !ok {
      return &ValidationError{Key: key, Err: "missing"}
    }
    if err := t.Validate(value); err != nil {
      return &ValidationError{Key: key, Err: err.Error()}
    }
  }
  return nil
}
```

### Phase 2: Trellis Integration (~1 day)

Extend existing components:

```go
// pkg/adapters/loam/metadata.go (ADD)
type NodeMetadata struct {
  // ... existing
  ContextSchema map[string]string  // "api_key" → "string"
}

// pkg/domain/node.go (ADD)
type Node struct {
  // ... existing
  ContextSchema schema.Schema  // Parsed, validated types
}

// internal/runtime/engine.go (EXTEND)
import "github.com/aretw0/trellis/pkg/schema"

func (e *Engine) Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error) {
  // ... existing validations ...
  
  // NEW: Type validation (alongside required_context)
  if err := e.validateContextTypes(node, state); err != nil {
    return nil, false, err
  }
  
  // ... continue
}

func (e *Engine) validateContextTypes(node *domain.Node, state *domain.State) error {
  if node.ContextSchema == nil {
    return nil
  }
  return schema.Validate(node.ContextSchema, state.Context)
}
```

### Phase 3: Metadata Parsing (Loam Adapter) (~0.5 day)

In `pkg/adapters/loam/loader.go`, parse schema strings to typed Schema:

```go
func (l *Loader) parseContextSchema(schemaMap map[string]string) (schema.Schema, error) {
  result := make(schema.Schema)
  for key, typeStr := range schemaMap {
    t, err := parseType(typeStr) // "string" → schema.String()
    if err != nil {
      return nil, fmt.Errorf("invalid type for key %s: %w", key, err)
    }
    result[key] = t
  }
  return result, nil
}

func parseType(typeStr string) (schema.Type, error) {
  switch typeStr {
  case "string":
    return schema.String(), nil
  case "int":
    return schema.Int(), nil
  case "float":
    return schema.Float(), nil
  case "bool":
    return schema.Bool(), nil
  case "[string]":
    return schema.Slice(schema.String()), nil
  // ... etc
  default:
    return nil, fmt.Errorf("unsupported type: %s", typeStr)
  }
}
```

### Phase 4: Documentation & Examples (~1 day)

- [docs/reference/node_syntax.md](../reference/node_syntax.md): Add `context_schema` section
- New example: `examples/typed-flow/` with success/failure cases
- [docs/TECHNICAL.md](../TECHNICAL.md): Section "Type Safety & Schema Validation"
- Test coverage in `internal/runtime/validation_test.go`

---

## 4. Example Usage

### Frontmatter Syntax

```yaml
---
id: checkout
required_context:
  - api_key
  - order_id
context_schema:
  api_key: string
  order_id: int
  items: [string]
---

# Checkout Flow

Processing order **{{ .order_id }}** with **{{ len(.items) }}** items.
```

### Runtime Validation

```go
// Valid
context := map[string]any{
  "api_key": "sk-123",
  "order_id": 42,
  "items": []string{"A", "B"},
}
// ✅ Passes schema.Validate

// Invalid - type mismatch
context := map[string]any{
  "api_key": 123,  // Expected string, got int
  "order_id": "wrong",  // Expected int, got string
}
// ❌ Returns ContextSchemaValidationError
```

---

## 5. Extraction Path (v0.8+ Option)

If evidence emerges of broader use (beyond Trellis):

- Pull `pkg/schema` into `../schema/` repo
- Update imports: `github.com/aretw0/trellis/pkg/schema` → `github.com/aretw0/schema`
- Decide on scope: `schema` (if grows beyond types) or `loam-schema` (if stays focused)
- Cost: ~4 hours refactoring across codebase

**Preconditions for extraction**:

- Trellis using schema actively (v0.7.5+)
- Evidence of need from another consumer OR internal Trellis growth (JSON Schema, coerção, etc)
- Clear scope definition

---

## 6. Risks & Mitigation

| Risk | Probability | Mitigation |
|------|-------------|-----------|
| Naming/scope unclear when extracting | Medium | Document current scope bounds. Revisit v0.8. |
| Schema impl locks us into design | Low | `Type` interface is simple, easy extend. |
| Code duplication if extracted | Low | Clean pkg boundary, trivial code movement. |

---

## 7. Timeline

- **Phase 1-2**: ~2 days
- **Phase 3**: ~0.5 day
- **Phase 4**: ~1 day
- **Total**: ~4 days target for v0.7.5 (can parallelize tests)

---

## 8. References

- [DECISIONS.md](../DECISIONS.md) - Decision log entry
- [PLANNING.md § v0.7.5](../PLANNING.md) - Feature tracking
- [Node Syntax Reference](../reference/node_syntax.md)
- [Loam Package](https://github.com/aretw0/loam)
