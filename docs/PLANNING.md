# Planning: Trellis

> Para filosofia e arquitetura, [consulte o README](../README.md).

## 1. Roadmap

### v0.7.14: The "Chat UI" Patch [CURRENT]

- [x] **Chat UI Polishing**: Evoluído o `reactivity-demo` para uma interface de chat web dedicada (integrada na CLI como `/ui`), com suporte robusto a SSE e auto-avanço de nós intermédios em background (`navigate("")`).
- [x] **Reactivity Hardening**: Implementados testes E2E headless rigorosos via `go-rod` e testes estressando o sistema SSE no backend com 100 eventos simultâneos por sessão.
- [x] **ToolResult via HTTP**: `Navigate` handler aceita `ToolResult` além de `string`. Frontend injeta resultado de ferramenta diretamente via API.
- [x] **Kitchen Sink Interpolation**: Nó `kitchen_sink_node` no fixture `ui_exhaustive` documenta e testa todos os padrões de interpolação suportados. Limitações mapeadas.
- [x] **Makefile**: Targets `make test-ui` e `make test-ui-headed` para rodar os testes E2E com ou sem browser visível.
- [x] **`mapStateToDomain` bidirecional**: `status` e `pending_tool_call` enviados pelo cliente são ignorados no parse. Necessário para retomada de sessão em `waiting_for_tool`. ✅ **FIXED**: Adicionado mapeamento bidirecional em `pkg/adapters/http/server.go`.

### 🩹 v0.7.15 (Patch): Chained Context Enforcement

**Focus**: Fix a pathological `context.Background()` detachment in `cmd/trellis/serve.go` identified during the lifecycle v1.7.1 ecosystem audit. The shutdown context for the HTTP server must respect the urgency escalation signalled by the parent lifecycle context (e.g., force-exit triggered by user mashing Ctrl+C).

- [ ] **`cmd/trellis/serve.go`**: Replace `context.WithTimeout(context.Background(), 5*time.Second)` with `context.WithTimeout(ctx, 5*time.Second)` in the HTTP server shutdown path.

### 🩹 v0.7.16 (Patch): Template Engine Hardening

**Foco**: Corrigir as limitações de interpolação identificadas pelo kitchen sink do v0.7.14 e tornar o `DefaultInterpolator` mais expressivo.

- [ ] **Server-side**: Avaliar se devemos facilitar o envio de markdown pré-renderizado para a interface web não aparecer markdown cru ou se isso deveria ser responsabilidade do cliente. O ideal é que o `DefaultInterpolator` seja capaz de lidar com os casos mais comuns (ex: `{{ default }}`) sem exigir FuncMap personalizada.
- [ ] **Client-side**: Implementar renderização de markdown no frontend para mensagens que contenham blocos de código ou formatação, garantindo que a UI seja amigável mesmo quando o backend não puder renderizar tudo perfeitamente.
- [ ] **FuncMap**: Registrar funções utilitárias no `template.New` em `internal/runtime/engine.go`: `default`, `index`, `toJson`, `coalesce`. Isso permite `{{ default "N/A" .missing_key }}` e acesso a campos de mapas dinâmicos.
- [ ] **`default_context` propagation**: Investigar por que o `default_context` definido em `start.md` não chega ao template. Verificar se o parser YAML do Loam faz merge correto no `domain.Context` antes da renderização.
- [ ] **`tool_result` typed access**: O resultado de ferramenta é armazenado como `interface{}` (struct interna `ToolResult{ID, Result}`). Avaliar se deve ser achatado (`map[string]any`) antes de ser salvo no contexto, possibilitando `{{ .tool_result.received }}`.
- [ ] **Documentar limitações**: Atualizar `docs/reference/node_syntax.md` com uma seção clara sobre o que o `DefaultInterpolator` suporta nativamente e o que requer FuncMap personalizada.
- [ ] **Testes**: Adicionar casos de teste unitários para o `DefaultInterpolator` cobrindo os padrões mais comuns e as limitações identificadas.
- [ ] **E2E Validation**: Validar que os fluxos existentes (ex: `ui_exhaustive`) continuam funcionando e que a UI renderiza mensagens corretamente após as mudanças.
- [ ] **Primitivas**: Considerar se é necessário adicionar primitivas de nó específicas para casos comuns de formatação ou manipulação de dados, aliviando a necessidade de lógica complexa no template.
- [ ] **Cognitive load**: Avaliar o que pode ser feito para reduzir a carga cognitiva por extrair lógica de formatação do template para o nó (ex: `type: format` com campos específicos).

### 📝 v0.7.17 (Patch): Documentation — Chat UI & Template Engine

**Foco**: Registrar formalmente o que foi implementado e as limitações descobertas durante o v0.7.14/v0.7.16. Toda documentação aqui dependente da estabilização do `DefaultInterpolator` (v0.7.16) antes de ser finalizada.

- [ ] **`docs/reference/node_syntax.md`**: Adicionar seção de limitações do `DefaultInterpolator` — o que funciona (`{{ .key }}`, `{{ if }}`, `{{ if eq }}`), o que não funciona sem FuncMap (`{{ default }}`), e o comportamento de `tool_result` como `interface{}`.
- [ ] **`docs/guides/frontend-integration.md`**: Expandir com guia completo do Chat UI embutido (`/ui`): como iniciar (`trellis serve`), fluxo de auto-advance, ciclo de vida do SSE, e como o cliente injeta `ToolResult`.
- [ ] **`docs/guides/running_http_server.md`**: Adicionar referência à UI embutida, aos endpoints `/ui`, `/navigate` com `ToolResult`, e ao campo `pending_tool_call` no schema de resposta.
- [ ] **`docs/TESTING.md`**: Documentar a estratégia de testes E2E com `go-rod`: targets do Makefile (`make test-ui`, `make test-ui-headed`), variável `TRELLIS_TEST_HEADLESS`, e o papel do fixture `ui_exhaustive` como contrato de comportamento da UI.

### 🏗️ v0.7.18: The "Automation" Patch

Foco: Melhorar a experiência de desenvolvimento e automação de scripts.

- [ ] **Single-File Execution** (ADR-0001): Oficializar suporte no `Runner` e na CLI para executar scripts definidos em arquivos únicos (`.yaml`, `.md`) sem exigir estrutura de diretórios (`trellis run my_script.md`).
- [ ] **Automation Nodes**: Testar integração nativa do Trellis com fluxos de Web Automation/Scraping (baseado nas lições do Wayang).

### 📦 v0.8: Ecosystem & Modularity (The "Mature" Phase)

Foco: Ferramentaria avançada e encapsulamento para grandes bases de código. Transformar Trellis em uma Plataforma.

- [ ] **Ecosystem Convergence (The "Lobster Way")**: Adaptação para modelos de pipelines tipados e resilientes.
  - [ ] **Project Definition**: Utilizar `loam` para carregar `trellis.yaml` (Manifest) e validar inputs/configs de forma unificada.
  - [ ] **Lifecycle Sinergy**:
    - [ ] **Supervisor Mount**: Tornar o Trellis um "Worker" compatível com o Supervisor do `lifecycle` (Gestão de Agentes).
    - [ ] **Unified Observability**: Integrar Introspecção (`State()`) e Telemetria (`pkg/metrics`) ao padrão do `lifecycle`.
    - [ ] **Durable Execution Delegation**: Depreciar `pkg/session` distribuído em favor do `lifecycle` agindo como Event Broker durável ([Ver ADR](architecture/durable-execution-delegation.md)).
  - [ ] **Resumable Protocols & Resilience**:
    - [ ] **Native Approval Gates**: Implementar `type: approval` com serialização de estado/token e Exit Code limpo (Safe Halt).
    - [ ] **Resume/Spawn Protocol**: Suporte a reidratação (`--resume <token>`) e contrato de mensagens para controle via Stdout.
    - [ ] **Native Retry Policies**: Formalizar suporte nativo a retentativas com *Exponential Backoff* direto no schema do nó (ex: `max_retries`, `backoff_strategy`).
- [ ] **SQLite Adapter**: Implementação de referência para `ports.StateStore` usando `database/sql`. Permite infraestrutura "Single-File" compartilhada com outras libs (ex: `whatsmeow`).
- [ ] **Developer Kit (SDK) & Type Safety**:
  - [ ] `trellis new flow` generators e scaffolding para facilitar o início.
  - [ ] **Type-Safe Context Engine**: Criar wrapper com Generics (ex: `trellis.NewGraph[MyContext]()`) para garantir que o state runtime seja formalmente tipado quando usado como biblioteca.
- [ ] **Trellis Gateway (Contextual Firewall)**: Implementação do "Hard Firewall" Stateful Proxy para controle de acesso dinâmico de ferramentas (MCP Router restrito por estados do DFA).
- [ ] **Language Server Protocol (LSP)**: Plugin de IDE (VSCode) para autocompletar nomes de nós, variáveis e ferramentas.
- [ ] **TUI Elements**: Widgets ricos para CLI (Select, MultiSelect, Password) via `charmbracelet/bubbles`.
- [ ] **Declarative Config (`trellis.yaml`)**: Permitir configurar Middlewares (Encryption, PII) e Adapters via arquivo de configuração.
  - *Refinement*: Internal middleware usage should be fully driven by this config.
- [ ] **WASM Target**: Compilar Trellis/Runner para WebAssembly, permitindo execução no Browser ou Edge (Cloudflare Workers).
- [ ] **gRPC Interface**: API binária para comunicação interna de baixa latência em malhas de serviço (Service Mesh).

### 🚀 v0.9: Compilers & Expressiveness (The "Expressive" Phase)

Foco: Melhorar a ergonomia e flexibilidade na construção de fluxos complexos baseando-se em conceitos abstratos.

- [ ] **Graph Compiler & Macro Nodes** (`type: flow`): Implementar a arquitetura abstrata (Lowering Phase) para compilar nós expressivos (estilo Colang) para o motor estrito do DFA sem penalidade em runtime, resolvendo o problema de verbosidade.
- [ ] **Advanced Validation**: Refatoramento de pré-flight checks e análise estática do compilador.

---

## 🌍 Ecosystem Evolution: Abstract Engine Architecture (2026+)

**Context**: O Trellis é uma **plataforma completa** (6 layers: UI, Protocols, Tooling, DSLs, Persistence, Engine). Para evitar perder insights valiosos durante a extração de componentes genéricos, adotamos uma estratégia incremental: **Refatorar → Validar → Extrair**.

**Vision**: Criar uma arquitetura onde DSLs (flow, life, scrape) compartilham um **Abstract Execution Engine** e componentes genéricos (protocols, persistence, tooling), mas mantêm suas próprias semânticas específicas.

**Reference**: Status e roadmap em [ECOSYSTEM_INTEGRATION.md](./ECOSYSTEM_INTEGRATION.md).

---

### 🔧 Phase 2a: Internal Restructuring (2-3 weeks) [NEXT]

**Objetivo**: Organizar o Trellis em camadas claras **sem extrair repositórios separados**. Manter 100% backward compatibility.

**Structure Target**:

```
trellis/pkg/
├── engine/         ← Core agnóstico (Node, Scheduler, State)
├── flow/           ← Flow-DSL específica (text, question, tool nodes)
├── protocols/      ← Candidato a extração (HTTP, MCP, SSE)
├── persistence/    ← Candidato a extração (StateStore, Session, SAGA)
├── tooling/        ← Candidato a extração (Tool Registry, Process Adapter)
├── ui/             ← Manter internamente (TUI, Web UI)
└── dsl/            ← Go builders (manter atual)
```

**Tasks**:

- [ ] Criar `pkg/engine/` com interfaces mínimas agnósticas
- [ ] Migrar lógica flow-specific para `pkg/flow/`
- [ ] Isolar HTTP/MCP/SSE em `pkg/protocols/`
- [ ] Isolar StateStore/Session em `pkg/persistence/`
- [ ] Isolar Tool Registry em `pkg/tooling/`
- [ ] Garantir 100% backward compat via facade `trellis.go`
- [ ] Validar: Todos os testes passam
- [ ] Validar: Arbour continua funcionando

**Blocker**: Node abstraction design decision (Hybrid recomendado).

**Deliverable**: Trellis refatorado internamente, código mais limpo, pronto para validação com life-dsl POC.

---

### 🧪 Phase 2b: Life-DSL POC (2-3 weeks)

**Objetivo**: Implementar **life-dsl dentro do repo Trellis** como experimento. Descobrir empiricamente o que é **realmente genérico**.

**Structure**:

```
trellis/pkg/
├── engine/    ← Life-dsl usa estas interfaces
├── flow/      ← Não modificar
└── life/      ← NOVO: Life-DSL implementation
    ├── types.go      (workers, schedules, health_checks)
    ├── compiler.go   (life.yaml → engine.Node)
    └── executors.go  (CLI, HTTP, Browser, Notify actions)

examples/life/
└── life.yaml
```

**Validation Questions** (empirical discovery):

- Life-dsl precisa de HTTP server? → protocols/ é genérico ✓
- Life-dsl precisa de MCP? → protocols/ é genérico ✓
- Life-dsl precisa de session management? → persistence/ é genérico ✓
- Life-dsl precisa de tool registry? → tooling/ é genérico ✓
- Life-dsl tem node types diferentes de flow? → engine/ precisa abstração maior

**Tasks**:

- [ ] Criar `pkg/life/` dentro do Trellis
- [ ] Parser de `life.yaml` (via Loam)
- [ ] Compiler `life.yaml → engine.Node`
- [ ] Executors (CLI, HTTP, Browser, Notify)
- [ ] Testar reuso: `pkg/protocols/`
- [ ] Testar reuso: `pkg/persistence/`
- [ ] Testar reuso: `pkg/tooling/`
- [ ] Documentar diferenças entre flow-dsl e life-dsl

**Blocker**: Phase 2a deve completar.

**Deliverable**: `examples/life/life.yaml` executando no Trellis. Lista concreta e validada do que pode ser extraído.

---

### 📦 Phase 2c: Surgical Extraction (1-2 weeks)

**Objetivo**: Extrair **apenas** o que foi **empiricamente validado** como genérico na Phase 2b.

**Extraction Candidates** (provisional):

1. **trellis-protocols** (HTTP, MCP, SSE) → usado por flow + life
2. **trellis-persistence** (StateStore, Session, SAGA) → usado por flow + life
3. **trellis-tooling** (Tool Registry, Process Adapter) → usado por todos

**NOT extracting** (yet):

- `pkg/engine/` — precisa maturar com 2+ DSLs
- `pkg/ui/` — themes são flow-specific, patterns precisam validação
- `pkg/dsl/` — cada DSL tem builders próprios

**Final Structure** (Phase 2c):

```
trellis/                  ← Flow DSL + Engine (monolítico ainda)
├── pkg/engine/
├── pkg/flow/
└── pkg/life/
└── go.mod
    depends on:
    - github.com/aretw0/trellis-protocols   (NEW)
    - github.com/aretw0/trellis-persistence (NEW)
    - github.com/aretw0/trellis-tooling     (NEW)
```

**Tasks**:

- [ ] Extrair `trellis-protocols` repo
- [ ] Extrair `trellis-persistence` repo
- [ ] Extrair `trellis-tooling` repo
- [ ] Atualizar Trellis `go.mod` para depender dos 3
- [ ] Atualizar Life-dsl para depender dos 3
- [ ] Criar `ECOSYSTEM_INTEGRATION.md` em cada repo
- [ ] Validar: Trellis funciona com deps externas
- [ ] Validar: Life-dsl funciona com deps externas
- [ ] Validar: Arbour continua funcionando

**Blocker**: Phase 2b deve identificar o que é genérico.

**Deliverable**: Ecosystem com shared components **validados empiricamente**.

---

### 🚀 Phase 3: Life-DSL Standalone Repo (2-3 weeks)

**Objetivo**: Criar repositório separado `aretw0/life-dsl`.

**Tasks**:

- [ ] Criar repo `aretw0/life-dsl`
- [ ] Migrar `trellis/pkg/life/` para novo repo
- [ ] Publicar life-dsl v0.1.0
- [ ] Documentar "Life as Code" use cases
- [ ] Integração com lifecycle v1.5
- [ ] Protocolo de remote agent (Git sync ou polling)

**Deliverable**: Life-dsl como projeto standalone, usando engine abstrato validado.

---

### 🎯 Phase 4: Extract Trellis-Engine (Future — 2027+)

**Aguardar**: Validação com **2+ DSLs** (flow, life) em produção.

**Então**: Extrair `trellis-engine` como core 100% genérico.

**Por enquanto**: Engine permanece dentro do Trellis (monolítico).

**Philosophy**: "Measure twice, cut once." — Validar empiricamente antes de abstrair.

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
