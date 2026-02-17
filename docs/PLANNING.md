# Planning: Trellis

> Para filosofia e arquitetura, [consulte o README](./README.md).

## 1. Roadmap

### ✅ v0.1: Bootstrap (MVP)

Foco: Estabelecer a arquitetura e provar o fluxo ponta-a-ponta.

- [x] **Core Domain**: Definição de `Node`, `Transition`, `State`.
- [x] **Engine**: Runtime básico de execução de passos.
- [x] **Loam Integration**: Uso de `TypedRepository` e normalização de markdown.
- [x] **Golden Path**: Verificação de robustez com "sujeira" e formatos mistos.
- [x] **CLI**: Interface básica funcional.

### ✅ v0.2: Stability & Certification

Foco: Integridade dos dados, testes e melhorias na CLI.

- [x] **Loam v0.8.3**: Suporte a formatos mistos (JSON/MD) e correção de retrieval.
- [x] **Certification Suite**: Testes de integração (TDD) para garantir corretude do Engine.
- [x] **CLI Hardening**: Lógica de saída genérica (Sink State) e supressão de output repetido.
- [x] **Project Cleanup**: Organização de exemplos e testes.
- [x] **Documentation Fix**: Corrigir links quebrados no README (Quick Start).

### ✅ v0.3: Developer Experience (DX) & Tooling

Foco: Ferramentas para quem *constrói* os fluxos (Toolmakers). Garantir confiança e testabilidade.

- [x] **Public Facade (Root Package)**: Refatorar para expor API no root `github.com/aretw0/trellis`.
- [x] **CLI & Runner Architecture**: Extrair loop para `Runner` e adotar `spf13/cobra` para gerenciar comandos (`run`, `graph`, `validate`).
- [x] **Compiler Validation**: O Compiler deve validar links mortos. (De-prioritized for CLI focus).
- [x] **Delegated Logic Integration**: Suporte a condicionais (`condition: is_vip`) e interpolação simples. A lógica real reside em callbacks no código Go (Host), não no Markdown.
- [x] **Introspection (Graphviz/Mermaid)**: Comando `trellis graph` para exportar a visualização do fluxo. "Documentation as Code".
- [x] **Headless Runner**: Capacidade de executar fluxos sem interface visual para testes automatizados de regressão.

### ✅ v0.3.1: Onboarding & Decoupling (The "Adoption" Phase)

Foco: Reduzir a barreira de entrada, clarificar a arquitetura para novos usuários e posicionar para a Era da IA.

- [x] **Loam Decoupling**: Refatorar `trellis.New` para tornar o Loam opcional via Functional Options pattern (`trellis.WithLoader`).
- [x] **MemoryLoader**: Implementar um adaptador `in-memory` oficial. Essencial para testes unitários de consumidores e para quem quer "hardcodar" o grafo em Go.
- [x] **Minimalist "Hello World"**: Criar `examples/hello-world` demonstrando o uso do `MemoryLoader` (sem arquivos, apenas Go).
- [x] **AI/LLM Documentation**: Adicionar seção "Trellis for AI Agents" no `PRODUCT.md` explicando o padrão "Symbolic/Neuro Architecture".
  - *Concept*: Trellis como "Deterministic Guardrails" para LLMs.
- [x] **Documentation Revamp**:
  - [x] Atualizar README: Diagrama "Host -> Trellis -> Adapter".
  - [x] Clarificar que Loam é "Batteries Included", mas opcional.

### ✅ v0.3.2: Reference Implementation (Minimal TUI)

Foco: Prover uma referência de implementação para TUI/SSH sem exageros. O objetivo é inspirar, não criar um framework de UI.

- [x] **Basic TUI Renderer**: Integração simples com `charmbracelet/glamour` apenas para sanitização e renderização básica de Markdown.
  - *Caveat (Resizing)*: O Renderer é inicializado uma única vez. Redimensionamento de janela durante a execução pode não atualizar o *word-wrapping* corretamente.
  - *Caveat (AutoStyle)*: `WithAutoStyle` depende do terminal reportar corretamente o fundo (Light/Dark). Pode falhar em certos terminais Windows, exigindo flag manual no futuro.
- [x] **Interactive Inputs Prototype**: PoC de como o Engine pode solicitar inputs complexos, delegando a UI para o Host.
  - *Constraint*: O Engine deve solicitar **dados** (ex: "OneOf: [A, B]"), não **widgets** (ex: "SelectBox"). Evitar acoplamento visual.
  - *Certification*: Adicionado `TestCertification_Examples` para validar a integridade dos exemplos públicos (`examples/tour`).
- [x] **Consolidate Examples**: Avaliar fusão de `interactive-demo` com `hello-world` para reduzir poluição na raiz.
  - *Action*: Renomeado `interactive-demo` para `low-level-api` e criado índice no `examples/README.md`.
- [x] **Dev Mode (Hot Reload)**: Implementar monitoramento de arquivos (Watch) via Loam.
  - *Estratégia*: Utilizar suporte nativo de `Watch` do Loam v0.9.0+.
  - *Caveat (State Handling)*: Não tentar reconciliação complexa de estado. Se o grafo mudar estruturalmente, reiniciar a sessão ou exibir aviso.
  - *Status*: Implementado `RunWatch` com tratamento de sinais e debounce.
- [x] **Documentation**: Guia explícito para "Interactive Inputs". O exemplo existe, mas falta documentação de referência.
- [x] **Hardening**: Testes de estresse para o Watcher (simular falhas de reload e múltiplos saves rápidos).

### ✅ v0.3.3: Stateless & Serverless (The "Cloud" Phase)

Foco: Preparar o Trellis para arquiteturas efêmeras (Lambda, Cloud Functions) típicas de Agentes de IA.

- [x] **Stateless Engine**: Garantir que as funções `Render` e `Navigate` sejam puramente funcionais.
- [x] **JSON IO**: Garantir que o runner possa operar puramente com Input JSON -> Output JSON, sem TTY.
- [x] **Validator Refactor**: Reimplementar `trellis validate` para operar sobre a abstração `GraphLoader`, permitindo validar grafos em memória ou bancos, não apenas arquivos.
- [x] **Strict Serialization**: Implementar suporte a `Strict Mode` global (Loam v0.10.4+). Garante consistência de tipos (`json.Number`) tanto para JSON quanto Markdown/YAML. (Regression test: `tests/serialization_test.go`).

### ✅ v0.4: Scale, Protocol & Integration (The "System" Phase)

Foco: Arquitetura para sistemas complexos, distribuídos e integração profunda com LLMs.

- [x] **Sub-Grafos (Namespaces)**: Capacidade de um nó apontar para outro arquivo/grafo (`jump_to: "modules/checkout/start"`). Permite modularização via diretórios e IDs implícitos.
- [x] **Stateless & Protocol Adapters**:
  - [x] **HTTP Server**: Adaptador JSON via `net/http`. [Veja o Guia](../docs/guides/running_http_server.md).
  - [x] **Server-Sent Events (SSE)**: Endpoint para notificar hot-reload em clientes web.
  - [x] **MCP Server (Model Context Protocol)**: Expor Trellis como ferramentas (`render`, `navigate`) e recursos (`graph`) para LLMs.
- [x] **Side-Effect Protocol (Tool Use)**: Padronização de como o Trellis solicita ações ao Host (Function Calling), alinhado com padrões de LLM (OpenAI Tool Spec).

### ✅ v0.4.1: Polimento & Extensibilidade

- [x] **Technical Debt & Hardening**:
  - [x] **System Messages**: Adicionar suporte a `IOHandler.SystemOutput` para separar mensagens de sistema do conteúdo.
  - [x] **Metadata-driven Safety**: Permitir `metadata.confirm_msg` para personalizar prompts do Middleware.
  - [x] **Interpolation Engine**: Substituir `strings.ReplaceAll` por template engine robusto (`Interpolator` Interface).
  - [x] **Async JSON Runner**: Refatorar `JSONHandler` para evitar bloqueio no Stdin (Event Loop).
  - [x] **OpenAPI Sync**: Garantir geração automatizada do código (oapi-codegen).
  - [x] **Refactoring: Terminology**: Renomear `State.Memory` para `State.Context` e `adapters/memory` para `adapters/inmemory`.
  - [x] **Refactoring: Legacy Cleanup**: Remover `memory_loader.go` antigo.
- [x] **Side-Effect Protocol Integration (Phase 2)**:
  - [x] **Tool Registry**: Implementar registro real de funções/scripts para evitar mocks.
  - [x] **Human-in-the-loop**: Implementado via `ConfirmationMiddleware`.
  - [x] **Loam Support**: Definir ferramentas em Markdown/Frontmatter.
  - [x] **Tool Libraries**: Suporte a referências de ferramentas (import) via chave polimórfica.
    - *Requirement*: Validar schema manualmente (`[]any`), detectar ciclos de importação e respeitar shadowing (local > import).

### 🧠 v0.5: Semantic Core (The "Pure" Phase)

Foco: Remover heurísticas de CLI do Core Engine e alinhar tipos de nós com semântica de State Machine pura.

- [x] **Non-Blocking Text**: Alterar semântica padrão de `type: text` para "Pass-through" (não bloqueia).
- [x] **Explicit Inputs**: Introduzir `type: prompt` ou `wait: true` para nós que exigem pausa/input.
- [x] **Data Binding (Input)**: Suporte a `save_to: "variable_name"` para salvar dados no `State.Context`.
- [x] **Context Namespacing**: Isolar variáveis de usuário (`user.*`) de variáveis de sistema (`sys.*`) para evitar Overwrite acidental.
- [x] **Lifecycle Cleanup**: Adotar padrão **Resolve** (Read Context, Deep Interpolation), **Execute** (Side-Effect), **Update** (Write Context).
- [x] **Type Erasure Fix**: Permitir que `save_to` armazene objetos complexos (`any`) de resultados de Tools, não apenas strings.
- [x] **Syntactic Sugar: Options**: Suporte a `options` como alias para `transitions` com `condition` implícita (Precedência: Options > Transitions).
- [x] **Syntactic Sugar: Root `to`**: Permitir `to: "next_node"` na raiz quando houver apenas uma transição incondicional (Menos verbosidade).
- [x] **Manual Migration**: Atualizar grafos de exemplo (`examples/`) para usar `wait: true` ou `type: prompt` onde necessário. (Análise: ~14 arquivos, inviável automação).

### 🛡️ v0.5.1: Robustness & Observation (The "Production" Patch)

Foco: Tornar o Trellis seguro e observável para rodar em produção.

- [x] **Error Handling**: Adicionar transição explícita `on_error: "node_id"` para recuperação automática de falhas em Tools. Implementada estratégia "Fail Fast" para erros não tratados.
- [x] **Observability Hooks**: Refatorar Engine para emitir eventos (`OnTransition`, `OnNodeEnter`) permitindo instrumentação externa (OpenTelemetry).
- [x] **Data Schema Validation**: Permitir definição de `required_context` no início do grafo para Fail Fast.

### 🛡️ v0.5.2: Control & Safety (The "Brakes" Phase)

Foco: Mecanismos de controle de execução e segurança. O Trellis deve ser interrompível e seguro por padrão, essencial para orquestração de Agentes IA imprevisíveis.

- [x] **Global Signals (Interrupts)**: Mecanismo nativo para lidar com sinais de interrupção (Ctrl+C, Timeout) e comandos globais ("cancel") convertendo-os em eventos de transição (`on_signal`).
- [x] **Graceful Shutdown**: Implementado `SignalManager` para garantir cancelamento limpo de contextos e `OnNodeLeave` hooks mesmo em interrupções forçadas.
- [x] **Input Sanitization**: Validar limitações físicas de input (tamanho, caracteres invisíveis) antes de injetar no State. Proteção contra DoS e contaminação de logs.

### ✅ v0.5.3: Signals & Developer Experience (The "Ergonomics" Patch)

Foco: Facilitar a vida de quem cria fluxos com Context Injection e melhor controle de sinais.

- [x] **Context Injection**: Adicionar flag `--context '{"key": "val"}'` à CLI para facilitar testes e integração.
- [x] **Default Context (Mocks)**: Permitir declarar valores padrão (`default_context`) no frontmatter para facilitar o desenvolvimento local e mocks de dependências.
- [x] **Global Signal Contexts**: Expandir `on_signal` para suportar `timeout` (System Signals) e `webhook` (External Signals).
- [x] **CLI DX**: Melhorias de output e logs para feedback mais limpo.

### ✅ v0.6: Integration & Persistence (The "Durable" Phase)

Foco: Transformar o Trellis de um Engine Stateless em uma solução de **Durable Execution** (inspirado em Temporal), permitindo fluxos de longa duração e recuperação de falhas.

- [x] **State Persistence Layer**: Definir interface `StateStore` (Load/Save/Delete) desacoplada do Core.
  - *Filosofia*: Snapshotting de Estado para permitir "Sleep & Resume" (Persistência, não Event Sourcing por enquanto).
- [x] **Adapters de Persistência**:
  - [x] **file.Store**: Persistência em JSON local. Permite "CLI Resumable" e debugging fácil.
  - [x] **Redis/Memory**: Interfaces de referência para alta performance.
- [x] **Runner Refactor**: Migrar `Runner` para Functional Options Pattern (remover `sessionID` de `Run`).
  - [x] **Session CLI**: Comandos para listar/inspecionar sessões (`trellis session ls`).
- [x] **Session Manager Pattern**: Implementação de referência para lidar com Concorrência (Locking) e ciclo de vida de sessão.
- [x] **SAGA Support (Compensation)**: Padrões e exemplos de como implementar transações compensatórias (`undo_action`) manuais.
  - [x] Example: `examples/manual-saga`
  - [x] Guide: `docs/guides/manual_saga_pattern.md`
  - *Caveat*: Atual implementação com `file.Store` segue modelo **Baton Passing** (Processo A para, Processo B continua). Não suporta "Remote Control" (Processo A acorda) sem polling/watch.
- [x] **Security Hooks**: Middlewares de persistência para Criptografia (Encryption at Rest) e Anonimização de PII no Contexto antes de salvar.
- [x] **Persistency Management (Chaos Control)**:
  - [x] **CLI**: `trellis session ls` (Listar), `rm` (Remover), `inspect` (Inspecionar State JSON).
  - [x] **Visual Debug**: `trellis graph --session <id>` para visualizar o "Caminho Percorrido" (Breadcrumbs) no diagrama (Overlay).
  - [x] **Auto-Pruning**: (Deferred to v0.7+) Documentado que a limpeza é responsabilidade do Admin (`trellis session rm`) para file.Store. Redis usa TTL nativo.
- [x] **Stateful Hot Reload (Live Coding)**:
  - [x] Permitir `--watch` e `--session` simultâneos.
  - [x] Ao recarregar o grafo, o Runner reidrata o estado da sessão existente, mantendo o histórico e variáveis.
  - [x] **Reload Guardrails**: Recuperação automática de Missing Node e Type Mismatch.
  - Permite corrigir typos e lógica sem reiniciar o fluxo do zero.
  - *Risk Check*: Se o nó atual for deletado, fallback para erro ou inicio.
- [x] **CLI Observability Strategy (DX)**:
  - [x] **Unified Logging**: Harmonizar output para Normal/Watch/Debug (Prefixos, Espaçamento).
  - [x] **Session UX**: Feedback explícito para eventos de Sessão (Start, Rehydrate, Reload).
  - [x] **Signal Handling**: Mensagens graciosas de "Interrupted" mascarando erros crus de Contexto.
  - [x] **Technical Debt (Backlog)**:
    - [x] `pkg/session`: Fix Lock Leaking (RefCounting) to prevent infinite growth.
    - [x] `internal/adapters/redis`: Add TTL Support (Expiration) for compliance.
    - [x] `internal/adapters/redis`: Optimize List implementation (Scan is O(N)).
    - [x] `internal/adapters/file_store`: Implement Atomic Writes (prevent corruption on crash).
    - [x] `pkg/runner`: Fix Non-Blocking text logic & Lifecycle consistency for terminal nodes.
    - [x] `pkg/persistence`: Refine internal usage of Middleware. (See v0.8 Declarative Config).
    - [x] `pkg/engine`: Validate Saga constraints in manual flows. (See v0.7 Native Saga).

### ✅ v0.7: Protocol & Scale (The "Network" Phase)

Foco: Expandir as fronteiras do Trellis para redes e alta escala (Distributed Systems).

- [x] **Distributed Locking**: Implementação de referência de `SessionManager` usando Redis/Etcd para clusters.
- [x] **Tool Idempotency**: Suporte a `idempotency_keys` para chamadas de ferramentas, garantindo segurança em retentativas (Network Flakes).
- [x] **Native SAGA Orchestration**: Engine capaz de fazer rollback automático (`undo`) lendo o histórico de execução (Stack Unwinding), eliminando a necessidade de wiring manual de cancelamento.
  - [x] *Validation*: Ensure Saga constraints are enforced (e.g., matching undo types).
- [x] **Universal Action Semantics ("Duck Typing")**: Remover a restrição de `type: tool`. Se um nó tem intenção de ação (`do`), ele executa. Unifica "Falar" e "Fazer" num único nó (Text + Action), reduzindo fadiga.
  - *Constraint*: `do` e `wait` (Input) são mutuamente exclusivos por enquanto.
- [x] **Syntactic Sugar: on_timeout**: Alias semântico para `on_signal["timeout"]`. Melhora a DX alinhando com `on_error`.
- [x] **Process Adapter (Scriptable Tools)**: Adaptador seguro para executar scripts locais (`.sh`, `.js`, `.py`, `.ps1`) via `tools.yaml`.
  - *Strategy*: Foco em "Polyglot Examples" para demonstrar o contrato Unix (Env/Stdin/Stdout) sem SDKs complexos por enquanto.
- [x] **Granular SSE Events**: (Moved to v0.7.1)
- [x] **MCP Advanced**: (Moved to v0.7.1)
- [x] **WASM/gRPC**: (Moved to v0.8)

### 🏗️ v0.7.1: Documentation & Installation (An "Polish" Patch)

Foco: Melhorias de documentação que não bloquearam o release v0.7.0, além de suporte a gerenciadores de pacotes.

- [x] **Installation Managers**: Suporte oficial a `scoop` (Windows) e `homebrew` (Linux/Mac).
- [x] **Architectural Decisions**: Extração do log de decisões para `DECISIONS.md` para manter `TECHNICAL.md` focado.
- [x] **GoDoc Server**: Ferramenta local para visualização de documentação de código.
- [x] **Documentation & Identity Polish**: Consolidação do README e **PRODUCT.md** com foco em "Neuro-Symbolic", "Resiliência" (SAGA) e limites do sistema (Constraints).

### ✅ v0.7.2: Ecosystem Unification (The "Core" Refactor)

Foco: Centralizar lógica repetitiva entre projetos do ecossistema (`trellis`, `tobot`, `fiscus`) para evitar duplicação e garantir consistência de comportamento (especialmente em Sinais e IO).

- [x] **Lifecycle Library**: Criação da lib `github.com/aretw0/lifecycle` para centralizar:
  - **SignalContext**: Lógica de duplo sinal (SIGINT vs SIGTERM).
  - **Terminal IO**: Abstração cross-platform (`CONIN$` no Windows) para leitura segura de input.
- [x] **Trellis Adoption**: Refatoração do Trellis para delegar essa responsabilidade à nova lib (Removed ~100 LOC).
- [x] **Dependency Switching**: Makefile targets (`use-local`, `use-pub`) para facilitar o desenvolvimento simultâneo de libs e cosumidores.

### ✅ v0.7.3: Polishing Lifecycle Synergy

Foco: Refinar o comportamento da CLI e ferramentas externas após a integração com a lib `lifecycle`.

- [x] **Input Goroutine Stability**: Corrigido vazamento de goroutines (`handleInput`) que causava "bloqueio" de input após interrupções (`Ctrl+C`).
- [x] **Tool Path Resolution**: Implementado `BaseDir` no `ProcessRunner`. Ferramentas externas (Scripts Python/Node) agora são resolvidas relativas ao diretório do fluxo, não do CWD.
- [x] **CLI Ergonomics**: Promoção de flags para o `rootCmd` e suporte a subcomando default. Permite rodar `trellis ./flow --debug` de forma intuitiva.
- [x] **Registry & Inline Unified**: Limpeza da lógica de carregamento de ferramentas e re-habilitação de logs de debug limpos.
- [x] **Atomic Commits**: Organização de todo o trabalho acumulado em 11 commits semânticos e atômicos.

### ✅ v0.7.4: Infrastructure & Interoperability

Foco: Estabilizar o ambiente de desenvolvimento e preparar a integração com ferramentas de diagnóstico.

- [x] **Dev Environment Interoperability**:
  - [x] **Cross-Platform Makefile**: Refatoração completa para suportar Windows e POSIX simultaneamente via GNU Make.
  - [x] **Go Workspace Sync**: Mecanismo de `DROP_WORK` com normalização de paths (`subst`) para garantir funcionamento cross-platform.
  - [x] **Dependency Automation**: Novos targets `work-on/off-[lib]` para `lifecycle`, `procio`, `loam` e `introspection`.
- [x] **Introspection Strategy Analysis**:
  - [x] **Technical Audit**: Análise de compatibilidade entre o gerador Mermaid interno e a lib `introspection`.
  - [x] **Strategy**: Manter visualização interna para grafos complexos; adotar `introspection` para snapshots de estado (v0.7.5).
- [x] **Lifecycle 1.5**: Avaliar se esta tudo estável para liberar a lifecycle ser publicada na v1.5.
  - **Verdict**: ✅ Estável. A suíte de testes passou (`make test`) utilizando as versões locais (`go.work`) das libs `lifecycle` (`main`), `procio` (`main`) e `introspection` (`main`). Nenhuma regressão detectada.
  - [x] **Release v1.5**: Publicar `lifecycle` v1.5.0 com breaking changes (SignalContext, Terminal IO) e atualizar dependências no `go.mod`.

### 🏗️ v0.7.5: Developer Experience & Type Safety (The "DX" Patch)

Foco: Trazer segurança de tipos e melhoria de experiência do desenvolvedor (DX).

- [x] **Lifecycle Workers & Supervisors**: Avaliar se o `trellis.Runner` deve ser implementado como um `Worker` nativo da lib `lifecycle` para melhor gestão de pools.
- [x] **Unified Observability**: Integrar a introspecção do Engine com os coletores de métricas e introspecção da lib `lifecycle`.
  - Implementar `Engine` como `introspection.TypedWatcher[EngineState]`
  - Usar `introspection.AggregateWatchers` para vista unificada (Engine + Workers + Signals)
  - ⚠️ **NÃO** usar introspection para geração de Mermaid (Performance). Manter o gerador interno para visualização de grafos complexos.
- [x] **Trellis as Lib (API Polish)**: Revisão da superfície pública (`pkg/runner`) para garantir que o Trellis seja tão fácil de usar como biblioteca quanto é como CLI.
- [ ] **Typed Flows**: Definição de schemas estritos para Contexto (`api_key: string`, `retries: int`), validados no carregamento e runtime. **Decision: Option A (Validation in Trellis) with Extraction Path**. See [docs/architecture/schema-validation-architecture.md](docs/architecture/schema-validation-architecture.md).
- [ ] **Go DSL / Builders**: Pacote `pkg/dsl` para construção de grafos Type-Safe em Go puro.
- [ ] **Granular SSE Events**: Update parcial de estado (Delta) para frontends reativos de alta performance.

### 📦 v0.8: Ecosystem & Modularity (The "Mature" Phase)

Foco: Ferramentaria avançada e encapsulamento para grandes bases de código. Transformar Trellis em uma Plataforma.

- [ ] **Ecosystem Convergence (The "Lobster Way")**: Adaptação para modelos de pipelines tipados e resilientes.
  - [ ] **Project Definition**: Utilizar `loam` para carregar `trellis.yaml` (Manifest) e validar inputs/configs de forma unificada.
  - [ ] **Lifecycle Sinergy**:
    - [ ] **Supervisor Mount**: Tornar o Trellis um "Worker" compatível com o Supervisor do `lifecycle` (Gestão de Agentes).
    - [ ] **Unified Observability**: Integrar Introspecção (`State()`) e Telemetria (`pkg/metrics`) ao padrão do `lifecycle`.
  - [ ] **Resumable Protocols**:
    - [ ] **Native Approval Gates**: Implementar `type: approval` com serialização de estado/token e Exit Code limpo (Safe Halt).
    - [ ] **Resume/Spawn Protocol**: Suporte a reidratação (`--resume <token>`) e contrato de mensagens para controle via Stdout.
- [ ] **SQLite Adapter**: Implementação de referência para `ports.StateStore` usando `database/sql`. Permite infraestrutura "Single-File" compartilhada com outras libs (ex: `whatsmeow`).
- [ ] **Developer Kit (SDK)**: `trellis new flow` generators e scaffolding para facilitar o início.
- [ ] **Language Server Protocol (LSP)**: Plugin de IDE (VSCode) para autocompletar nomes de nós, variáveis e ferramentas.
- [ ] **TUI Elements**: Widgets ricos para CLI (Select, MultiSelect, Password) via `charmbracelet/bubbles`.
- [ ] **Declarative Config (`trellis.yaml`)**: Permitir configurar Middlewares (Encryption, PII) e Adapters via arquivo de configuração.
  - *Refinement*: Internal middleware usage should be fully driven by this config.
- [ ] **WASM Target**: Compilar Trellis/Runner para WebAssembly, permitindo execução no Browser ou Edge (Cloudflare Workers).
- [ ] **gRPC Interface**: API binária para comunicação interna de baixa latência em malhas de serviço (Service Mesh).

---

## 2. Breaking Changes & Versioning Strategy {#breaking-changes-141-150}

### Estratégia de Versionamento

O Trellis adota **Semantic Versioning** (SemVer) dentro da **série v1.x**. Isto significa:

- **v1.0.0 → v1.x.y**: Mudanças backwards-compatible (novos recursos, patches).
- **v1.x.0 → v1.(x+1).0**: Podem incluir breaking changes **documentados**, mas o module path permanece `github.com/aretw0/trellis`.

> **Decisão sobre v2**: Para evitar a fadiga de migração de módulos Go (que requereria `github.com/aretw0/trellis/v2` no `go.mod`), optamos por **permanecer na v1** durante todo o lifecycle principal do projeto. Breaking changes significativos serão documentados explicitamente entre minor versions.

### Breaking Changes: v1.4.1 → v1.5.0

A versão **v1.5.0** introduz mudanças significativas na arquitetura de gerenciamento de ciclo de vida e IO:

#### 🔄 Lifecycle Router (Signals & Input Unification)

**Antes (≤ v1.4.1)**:

- O `Runner` capturava sinais POSIX (`SIGINT`, `SIGTERM`) diretamente.
- A leitura de input (`Stdin`) era bloqueante e tratada no loop principal.
- Diferentes estratégias entre plataformas (Windows vs POSIX).

**Depois (≥ v1.5.0)**:

- Introdução da biblioteca externa **`github.com/aretw0/lifecycle`**.
- O `Runner` delega captura de sinais e input para o **Lifecycle Router**.
- Uso de `signal.Context` do `lifecycle` para tratamento unificado cross-platform.
- Input é consumido via eventos roteados, desacoplando do loop de execução.

**Impacto de Migração**:

- **Consumidores da CLI**: Sem mudanças visíveis ao usuário final.
- **Library Users**: Se você instancia `Runner` diretamente em Go, pode ser necessário ajustar a inicialização de contextos. Consulte exemplos atualizados em `examples/low-level-api`.

#### 📦 Dependências Externas

A integração com `lifecycle` introduz novas dependências:

- `github.com/aretw0/lifecycle` (v1.5.0+)
- `github.com/aretw0/procio` (transitivo)

**Recomendação**: Rode `go mod tidy` após atualizar para v1.5.0.

#### 🧪 Guia de Migração

Para projetos que usam Trellis como biblioteca:

```go
// ANTES (v1.4.1)
runner := runner.New(
    runner.WithEngine(engine),
    runner.WithHandler(ioHandler),
)

// DEPOIS (v1.5.0)
ctx := lifecycle.SignalContext(context.Background())
runner := runner.New(
    runner.WithEngine(engine),
    runner.WithHandler(ioHandler),
)
runner.Run(ctx) // Passa o contexto gerenciado
```

Consulte a documentação completa em [TECHNICAL.md](TECHNICAL.md#9-arquitetura-do-runner--io) para detalhes sobre a nova arquitetura.

---

> **Arquitetura & Decisões**: O histórico de decisões arquiteturais foi movido para [DECISIONS.md](./DECISIONS.md).

---
