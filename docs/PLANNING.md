# Planning: Trellis

> Para filosofia e arquitetura, [consulte o README](../README.md).

## 1. Roadmap

### v0.7.14: The "Chat UI" Patch [RELEASED]

- [x] **Chat UI Polishing**: Evoluído o `reactivity-demo` para uma interface de chat web dedicada (integrada na CLI como `/ui`), com suporte robusto a SSE e auto-avanço de nós intermédios em background (`navigate("")`).
- [x] **Reactivity Hardening**: Implementados testes E2E headless rigorosos via `go-rod` e testes estressando o sistema SSE no backend com 100 eventos simultâneos por sessão.
- [x] **ToolResult via HTTP**: `Navigate` handler aceita `ToolResult` além de `string`. Frontend injeta resultado de ferramenta diretamente via API.
- [x] **Kitchen Sink Interpolation**: Nó `kitchen_sink_node` no fixture `ui_exhaustive` documenta e testa todos os padrões de interpolação suportados. Limitações mapeadas.
- [x] **Makefile**: Targets `make test-ui` e `make test-ui-headed` para rodar os testes E2E com ou sem browser visível.
- [x] **`mapStateToDomain` bidirecional**: `status` e `pending_tool_call` enviados pelo cliente são ignorados no parse. Necessário para retomada de sessão em `waiting_for_tool`. ✅ **FIXED**: Adicionado mapeamento bidirecional em `pkg/adapters/http/server.go`.

### 🩹 v0.7.15 (Patch): Chained Context Enforcement

**Status**: [COMPLETED]

**Focus**: Fix a pathological `context.Background()` detachment in `cmd/trellis/serve.go` identified during the lifecycle v1.7.1 ecosystem audit. The shutdown context for the HTTP server must respect the urgency escalation signalled by the parent lifecycle context (e.g., force-exit triggered by user mashing Ctrl+C).

- [x] **`cmd/trellis/serve.go`**: Replace `context.WithTimeout(context.Background(), 5*time.Second)` with `context.WithTimeout(ctx, 5*time.Second)` in the HTTP server shutdown path.
- [x] **`pkg/adapters/process/resilience_test.go`**: Add Windows platform-specific assertions to account for limited signal propagation in background processes on Windows. The grace period verification now uses `runtime.GOOS` checks to differentiate behavior: Unix expects full 5s grace period, Windows allows force-kill within 10s.

### 🩹 v0.7.16 (Patch): Template Engine Hardening

**Foco**: Corrigir as limitações de interpolação identificadas pelo kitchen sink do v0.7.14 e tornar o `DefaultInterpolator` mais expressivo.

- [ ] **FuncMap**: Registrar funções utilitárias no `template.New` em `internal/runtime/engine.go`: `default`, `index`, `toJson`, `coalesce`. Isso permite `{{ default "N/A" .missing_key }}` e acesso a campos de mapas dinâmicos.
- [ ] **`default_context` propagation**: Investigar por que o `default_context` definido em `start.md` não chega ao template. Verificar se o parser YAML do Loam faz merge correto no `domain.Context` antes da renderização.
- [ ] **`tool_result` typed access**: O resultado de ferramenta é armazenado como `interface{}` (struct interna `ToolResult{ID, Result}`). Avaliar se deve ser achatado (`map[string]any`) antes de ser salvo no contexto, possibilitando `{{ .tool_result.received }}`.
- [ ] **Testes**: Adicionar casos de teste unitários para o `DefaultInterpolator` cobrindo os padrões mais comuns e as limitações identificadas.

**Escopo opcional (se sobrar tempo)**:

- [ ] **Server-side Markdown**: Avaliar envio de markdown pré-renderizado pelo backend vs responsabilidade do cliente.
- [ ] **Client-side Markdown**: Melhorar renderização de blocos/código no frontend.
- [ ] **Primitivas de formatação**: Explorar nós específicos (ex: `type: format`) para reduzir lógica em template.

### 📝 v0.7.17 (Patch): Documentation — Chat UI & Template Engine

**Foco**: Registrar formalmente o que foi implementado e as limitações descobertas durante o v0.7.14/v0.7.16. Toda documentação aqui dependente da estabilização do `DefaultInterpolator` (v0.7.16) antes de ser finalizada.

- [ ] **`docs/reference/node_syntax.md`**: Adicionar seção de limitações do `DefaultInterpolator` — o que funciona (`{{ .key }}`, `{{ if }}`, `{{ if eq }}`), o que não funciona sem FuncMap (`{{ default }}`), e o comportamento de `tool_result` como `interface{}`.
- [ ] **`docs/guides/frontend-integration.md`**: Expandir com guia completo do Chat UI embutido (`/ui`): como iniciar (`trellis serve`), fluxo de auto-advance, ciclo de vida do SSE, e como o cliente injeta `ToolResult`.
- [ ] **`docs/guides/running_http_server.md`**: Adicionar referência à UI embutida, aos endpoints `/ui`, `/navigate` com `ToolResult`, e ao campo `pending_tool_call` no schema de resposta.
- [ ] **`docs/TESTING.md`**: Documentar a estratégia de testes E2E com `go-rod`: targets do Makefile (`make test-ui`, `make test-ui-headed`), variável `TRELLIS_TEST_HEADLESS`, e o papel do fixture `ui_exhaustive` como contrato de comportamento da UI.

### 🏗️ v0.7.18: The "Automation" Patch

Foco: Melhorar a experiência de desenvolvimento e automação de scripts.

- [ ] **Single-File Execution** (ADR-0001): Oficializar suporte no `Runner` e na CLI para executar scripts definidos em arquivos únicos (`.yaml`, `.md`) sem exigir estrutura de diretórios (`trellis run my_script.md`).

- [ ] **Automation Nodes Spike**: Testar integração de automação web inspirada no Wayang, com foco em validar semântica mínima (`navigate`, `act`, `extract`, `paginate`) no motor atual.

### 📦 v0.8 (Deferred): Ecosystem & Modularity

**Foco**: Consolidar o que foi validado empiricamente antes de extrair componentes.

- [ ] **Project Definition**: `trellis.yaml` (manifest unificado via Loam)
- [ ] **Lifecycle Synergy**: Supervisor mount + observabilidade unificada + durable delegation
- [ ] **Resilience Primitives**: approval, resume/spawn, retry policies
- [ ] **SQLite Adapter**: referência para `ports.StateStore`
- [ ] **SDK / DX**: `trellis new flow` e melhorias de onboarding

---

## 🌱 Discovery Track (Paralela e Leve)

Objetivo: cultivar os DSLs sem reescrita da engine e sem extração prematura.

### Execution Contract v0 (comum aos DSLs)

- [x] **Node Abstraction Decision**: **Hybrid** (Interface core + Function adapters + Builder ergonomics)
- [ ] Definir contrato mínimo de execução: `state`, `status`, `context`, `events`, `metadata`
- [ ] Garantir extensibilidade por DSL em `metadata` (sem quebrar compat)
- [ ] Permitir executores híbridos: `flow`, `cli`, `browser`, `http`, `notify`

### Life-DSL (POC interno no Trellis)

- [ ] Criar `examples/life/life.yaml` como contrato mínimo (workers, triggers, health, persistence, notifications)
- [ ] Provar execução híbrida de `sub_workers` (`flow` + scripts/tools)
- [ ] Modelar políticas declarativas de falha/escalonamento por configuração

### Scrape-DSL (porta aberta)

- [ ] Criar spike inspirado no Wayang para validar DSL declarativo de scraping sobre o motor atual
- [ ] Definir capability flag para scheduler especializado no futuro (`scheduler: default|scrape`)
- [ ] Documentar critérios para evoluir scheduler dedicado (sem compromisso imediato de implementação)

### Gates de Decisão para Extração

- [ ] Gate 1: 2 DSLs funcionando no mesmo contrato por pelo menos 1 ciclo de release
- [ ] Gate 2: componentes compartilhados comprovados em runtime real (não só design)
- [ ] Gate 3: migração sem regressão para Trellis e Arbour

> **Regra**: extração acontece quando o uso provar; não antes.

---

## 🌍 Ecosystem Evolution (Pragmático)

**Context**: O Trellis é uma plataforma (UI, Protocols, Tooling, DSLs, Persistence, Engine), mas o roadmap agora prioriza entrega incremental e validação empírica.

**Reference**: Status de integração em [ECOSYSTEM_INTEGRATION.md](./ECOSYSTEM_INTEGRATION.md).

---

## Ecosystem Readiness (Lifecycle v1.8+)

Objetivo: Alinhar o Trellis com as primitivas de "Durable Execution" e sincronicidade do `lifecycle` v1.8+.

- [ ] **Barrier Primitives (v1.8)**:
  - [ ] Adotar as novas barreiras de sincronização do `lifecycle` para o orquestrador de passos paralelos, simplificando o controle de "Join" no motor.
- [ ] **Durable Event Router (v1.8)**:
  - [ ] Integrar os novos "Durable Sinks" do Router para persistência nativa de breadcrumbs e transições, reduzindo o acoplamento com o `pkg/session`.
- [ ] **Case Study Flagship (v1.9)**:
  - [ ] Curar snippets de código "Antes vs Depois" da implementação do `lifecycle` para o marketing oficial da lib (Case Study: "Trellis: De 500 linhas de sinal para 50").

---

> **Arquitetura & Decisões**: O histórico de decisões arquiteturais foi movido para [docs/architecture/HISTORY.md](./architecture/HISTORY.md).
