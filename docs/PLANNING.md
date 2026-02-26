# Planning: Trellis

> Para filosofia e arquitetura, [consulte o README](../README.md).

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

### ✅ v0.7.5: Lifecycle & Observability [COMPLETED]

Foco: Trazer segurança de tipos e melhoria de experiência do desenvolvedor (DX).

- [x] **Lifecycle Workers & Supervisors**: Avaliar se o `trellis.Runner` deve ser implementado como um `Worker` nativo da lib `lifecycle` para melhor gestão de pools.
- [x] **Unified Observability**: Integrar a introspecção do Engine com os coletores de métricas e introspecção da lib `lifecycle`.
  - [x] Implementar `Engine` como `introspection.TypedWatcher[EngineState]`
  - [x] Usar `introspection.AggregateWatchers` para vista unificada (Engine + Workers + Signals)
  - [x] ⚠️ **NÃO** usar introspection para geração de Mermaid (Performance). Manter o gerador interno para visualização de grafos complexos.
- [x] **Trellis as Lib (API Polish)**: Revisão da superfície pública (`pkg/runner`) para garantir que o Trellis seja tão fácil de usar como biblioteca quanto é como CLI.

### ✅ v0.7.6: Type Safety & Schema Validation (The "Contracts" Patch) [COMPLETED]

Foco: Segurança de tipos para definição de grafos.

- [x] **Typed Flows**: Definição de schemas estritos para Contexto (`api_key: string`, `retries: int`), validados no carregamento e runtime. **Decision: Option A (Validation in Trellis) with Extraction Path**. See [docs/architecture/schema-validation.md](architecture/schema-validation.md).
  - [x] **Core Schema Package**: `pkg/schema/` com Type interface (string, int, float, bool, array, custom).
  - [x] **Loam Adapter Integration**: Parse `context_schema` frontmatter e validar tipos em runtime.
  - [x] **Error Handling**: `ContextSchemaValidationError` com diagnostics claros.
  - [x] **Documentation & Examples**: `examples/typed-flow/`, atualizar `docs/reference/node_syntax.md`.

### ✅ v0.7.7: Type-Safe Builders (The "Foundations" Patch) [COMPLETED]

Foco: API Go inicial para construir grafos sem YAML/JSON.

- [x] **Go DSL / Builders**: Pacote `pkg/dsl` para construção de grafos Type-Safe em Go puro.
  - [x] **Fluent Builder API**: `dsl.New().Add("start").Text("...").Go("next")`.
  - [x] **MemoryLoader Integration**: Compilação direta para loader em memória.
  - [x] **Documentation & Examples**: `examples/dsl-graph/`, guia inicial em `docs/guides/building-graphs-go.md`.

### ✅ v0.7.8: Fluent API Completion & Documentation [COMPLETED]

Foco: Completar a DSL com suporte a ferramentas, SAGA e documentação técnica detalhada.

- [x] **Tool & SAGA Support**:
  - [x] **Tool Registration**: `Do(name, args)` e `Tools(tools...)` no `NodeBuilder`.
  - [x] **SAGA Support**: `Undo(name, args)` para transações compensatórias.
  - [x] **Terminal Nodes**: `Terminal()` alias para nós de saída.
- [x] **Advanced Testing Helpers**: DSL otimizada para asserções em testes unitários.
- [x] **Documentation & Diagrams**:
  - [x] Atualizar `docs/guides/building-graphs-go.md` com exemplos de ferramentas.
  - [x] Atualizar `pkg/dsl/doc.go` com a API correta.
  - [x] Adicionar diagrama de sequência da construção do grafo.

### ✅ v0.7.9: Real-Time Updates (The "Reactivity" Patch) [COMPLETED]

Foco: Atualizações parciais de estado para frontends reativos.

- [x] **Granular SSE Events**: Update parcial de estado (Delta) para frontends reativos de alta performance.
  - [x] **State Diff**: Detectar mudanças apenas em campos alterados (not full snapshot).
  - [x] **SSE Delta Protocol**: Serializar deltas em JSON compacto.
  - [x] **HTTP Server Integration**: Endpoint `/events` com suporte a filtering (ex: `?watch=context,history`).
  - [x] **Documentation & Examples**: Exemplo React/vanilla JS que consome deltas, guia em `docs/guides/frontend-integration.md`.
  - [x] **Technical Debt**:
    - [x] **Context Deletion Protocol**: Definir padrão para remover chaves do contexto (ex: `null`). [COMPLETED]
    - [x] **Default Signal Handlers (Proposal)**: Permitir configurar `on_signal_default` no nível do grafo.
    - [x] **SSE Tests Data Race**: Corrigir condição de corrida detectada pelo `-race`.

### ✅ v0.7.10: The "Signal" Patch [COMPLETED]

Foco: Consolidar a arquitetura de sinais e centralizar os schemas de resposta.

- [x] **Schema Centralization**: Unificação de `RenderResponse` e `RichResponse` no OpenAPI e adaptadores.
- [x] **Signal Centralization**: Refatoração da lógica de sinal para o `Runner` (centralizado via `lifecycle`).
- [x] **Default Handlers**: Implementar `on_signal_default` (restrito ao nó de entrada/root).
- [x] **Graceful Session Shutdown**: Garantir que o encerramento de uma sessão libere recursos (SSE, memória).
- [x] **Termination Logic Fix**: Corrigido bug onde o engine precisava de uma interação extra para detectar fim de fluxo.
- [x] **Warnings**: Sistema de logs avisando sobre configurações ignoradas em nós não-root.

### ✅ v0.7.11: The "Context" Patch [COMPLETED]

- [x] **Deletion Support**: Implementar o protocolo de deleção no `StateDiff` e no `Subscriber`.
- [x] **Efficiency**: Otimizar serialização de deltas grandes.

### ✅ v0.7.12: The "Structure" Patch [COMPLETED]

- [x] **Entrypoint fallback**: Suportar `start.md`, `main.md`, `index.md`, e `NomeDaPasta.md`.
- [x] **ID Collisions**: Detectar e reportar colisão de IDs em sub-grafos.

### ✅ v0.7.13: The "Ecosystem" Patch [COMPLETED]

Foco: Integrar as melhorias mais recentes nas bibliotecas fundamentais do ecossistema.

- [x] **Lifecycle Update**: Atualizar para a versão contendo `StopAndWait(ctx)` no `ProcessWorker` e simplificar a mecânica de teardown manual atual no `trellis/pkg/adapters/process/runner.go`.
- [x] **Loam Update**: Avaliar e aplicar as mais recentes atualizações de parser e schema do `loam` no Trellis para manter paridade e corrigir débitos.

### ✅ v0.7.14: The "Chat UI" Patch [COMPLETED]

- [x] **Chat UI Polishing**: Evoluído o `reactivity-demo` para uma interface de chat web dedicada (integrada na CLI como `/ui`), com suporte robusto a SSE e auto-avanço de nós intermédios em background (`navigate("")`).
- [x] **Reactivity Hardening**: Implementados testes E2E headless rigorosos via `go-rod` e testes estressando o sistema SSE no backend com 100 eventos simultâneos por sessão.
- [x] **ToolResult via HTTP**: `Navigate` handler aceita `ToolResult` além de `string`. Frontend injeta resultado de ferramenta diretamente via API.
- [x] **Kitchen Sink Interpolation**: Nó `kitchen_sink_node` no fixture `ui_exhaustive` documenta e testa todos os padrões de interpolação suportados. Limitações mapeadas.
- [x] **Makefile**: Targets `make test-ui` e `make test-ui-headed` para rodar os testes E2E com ou sem browser visível.

### 🩹 v0.7.15 (Patch): Chained Context Enforcement

**Focus**: Fix a pathological `context.Background()` detachment in `cmd/trellis/serve.go` identified during the lifecycle v1.7.1 ecosystem audit. The shutdown context for the HTTP server must respect the urgency escalation signalled by the parent lifecycle context (e.g., force-exit triggered by user mashing Ctrl+C).

- [ ] **`cmd/trellis/serve.go`**: Replace `context.WithTimeout(context.Background(), 5*time.Second)` with `context.WithTimeout(ctx, 5*time.Second)` in the HTTP server shutdown path.

### 🩹 v0.7.16 (Patch): Template Engine Hardening

**Foco**: Corrigir as limitações de interpolação identificadas pelo kitchen sink do v0.7.14 e tornar o `DefaultInterpolator` mais expressivo.

- [ ] **FuncMap**: Registrar funções utilitárias no `template.New` em `internal/runtime/engine.go`: `default`, `index`, `toJson`, `coalesce`. Isso permite `{{ default "N/A" .missing_key }}` e acesso a campos de mapas dinâmicos.
- [ ] **`default_context` propagation**: Investigar por que o `default_context` definido em `start.md` não chega ao template. Verificar se o parser YAML do Loam faz merge correto no `domain.Context` antes da renderização.
- [ ] **`tool_result` typed access**: O resultado de ferramenta é armazenado como `interface{}` (struct interna `ToolResult{ID, Result}`). Avaliar se deve ser achatado (`map[string]any`) antes de ser salvo no contexto, possibilitando `{{ .tool_result.received }}`.
- [ ] **`mapStateToDomain` bidirecional**: `status` e `pending_tool_call` enviados pelo cliente são ignorados no parse. Necessário para retomada de sessão em `waiting_for_tool`.

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
