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
- [ ] **Syntactic Sugar: Root `to`**: Permitir `to: "next_node"` na raiz quando houver apenas uma transição incondicional (Menos verbosidade).
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

### 🧩 v0.6: Integration & Persistence (The "Durable" Phase)

Foco: Transformar o Trellis de um Engine Stateless em uma solução de **Durable Execution** (inspirado em Temporal), permitindo fluxos de longa duração e recuperação de falhas.

- [x] **State Persistence Layer**: Definir interface `StateStore` (Load/Save/Delete) desacoplada do Core.
  - *Filosofia*: Snapshotting de Estado para permitir "Sleep & Resume" (Persistência, não Event Sourcing por enquanto).
- [x] **Adapters de Persistência**:
  - [x] **FileStore**: Persistência em JSON local. Permite "CLI Resumable" e debugging fácil.
  - [x] **Redis/Memory**: Interfaces de referência para alta performance.
- [x] **Runner Refactor**: Migrar `Runner` para Functional Options Pattern (remover `sessionID` de `Run`).
  - [x] **Session CLI**: Comandos para listar/inspecionar sessões (`trellis session ls`).
- [x] **Session Manager Pattern**: Implementação de referência para lidar com Concorrência (Locking) e ciclo de vida de sessão.
- [x] **SAGA Support (Compensation)**: Padrões e exemplos de como implementar transações compensatórias (`undo_action`) manuais.
  - [x] Example: `examples/manual-saga`
  - [x] Guide: `docs/guides/manual_saga_pattern.md`
- [x] **Security Hooks**: Middlewares de persistência para Criptografia (Encryption at Rest) e Anonimização de PII no Contexto antes de salvar.
- [x] **Persistency Management (Chaos Control)**:
  - [x] **CLI**: `trellis session ls` (Listar), `rm` (Remover), `inspect` (Inspecionar State JSON).
  - [x] **Visual Debug**: `trellis graph --session <id>` para visualizar o "Caminho Percorrido" (Breadcrumbs) no diagrama (Overlay).
  - [x] **Auto-Pruning**: (Deferred to v0.7+) Documentado que a limpeza é responsabilidade do Admin (`trellis session rm`) para FileStore. Redis usa TTL nativo.
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
  - [ ] **Technical Debt (Backlog)**:
    - [ ] `pkg/session`: Fix Lock Leaking (LRU/GC) for High-Traffic scenarios.
    - [x] `internal/adapters/redis`: Add TTL Support (Expiration) for compliance.
    - [x] `internal/adapters/redis`: Optimize List implementation (Scan is O(N)).
    - [x] `internal/adapters/file_store`: Implement Atomic Writes (prevent corruption on crash).
    - [x] `pkg/runner`: Fix Non-Blocking text logic & Lifecycle consistency for terminal nodes.
    - [ ] `pkg/persistence`: Refine internal usage of Middleware. (See v0.8 Declarative Config).
    - [ ] `pkg/engine`: Validate Saga constraints in manual flows. (See v0.7 Native Saga).

### 🔌 v0.7: Protocol & Scale (The "Network" Phase)

Foco: Expandir as fronteiras do Trellis para redes e alta escala (Distributed Systems).

- [ ] **Distributed Locking**: Implementação de referência de `SessionManager` usando Redis/Etcd para clusters.
- [ ] **Tool Idempotency**: Suporte a `idempotency_keys` para chamadas de ferramentas, garantindo segurança em retentativas (Network Flakes).
- [ ] **Native SAGA Orchestration**: Engine capaz de fazer rollback automático (`undo_action`) lendo o histórico de execução (Stack Unwinding), eliminando a necessidade de wiring manual de cancelamento.
- [ ] **Granular SSE Events**: Update parcial de estado (Delta) para frontends reativos de alta performance.
- [ ] **Process Adapter (Scriptable Tools)**: Adaptador genérico para executar scripts locais (`.sh`, `.js`, `.py`) ou Binários (Lambdas) como Ferramentas, sem recompilar o Runner. "Unix Philosophy".
- [ ] **MCP Advanced**: Suporte a Prompts (Templates gerenciados), Sampling (controle de custos) e Docker Containerized Tools.
- [ ] **WASM Target**: Compilar Trellis/Runner para WebAssembly, permitindo execução no Browser ou Edge (Cloudflare Workers).
- [ ] **gRPC Interface**: API binária para comunicação interna de baixa latência em malhas de serviço (Service Mesh).

### 📦 v0.8: Ecosystem & Modularity (The "Mature" Phase)

Foco: Ferramentaria avançada e encapsulamento para grandes bases de código. Transformar Trellis em uma Plataforma.

- [ ] **Module Encapsulation**: Escopo privado e contratos de entrada/saída para criar bibliotecas de nós reutilizáveis.
- [ ] **Typed Flows**: Definição de schemas estritos para Contexto (`api_key: string`, `retries: int`), validados no carregamento.
- [ ] **Developer Kit (SDK)**: `trellis new flow` generators e scaffolding para facilitar o início.
- [ ] **Language Server Protocol (LSP)**: Plugin de IDE (VSCode) para autocompletar nomes de nós, variáveis e ferramentas.
- [ ] **Go DSL / Builders**: Pacote `pkg/dsl` para construção de grafos Type-Safe em Go puro.
- [ ] **TUI Elements**: Widgets ricos para CLI (Select, MultiSelect, Password) via `charmbracelet/bubbles`.
- [ ] **Declarative Config (`trellis.yaml`)**: Permitir configurar Middlewares (Encryption, PII) e Adapters via arquivo de configuração, eliminando a necessidade de código Go (`main.go`) para setups padrão.

---

## 2. Decisões Arquiteturais (Log)

- **2025-12-11**: *Presentation Layer Responsibility*. Decidido que a limpeza de whitespace (sanitização de output) é responsabilidade da camada de apresentação (CLI), não do Storage (Loam) ou do Domain (Engine).
- **2025-12-11**: *Loam Integration*. Adotado `TypedRepository` para mapear frontmatter automaticamente, tratando o Loam como fonte da verdade para formatos.
- **2025-12-13**: *Logic Decoupling*. Adotada estratégia de "Delegated Logic". O Markdown declara *intenções* de lógica, o Host implementa.
- **2025-12-13**: *Encapsulation*. `NodeMetadata` e `LoaderTransition` mantidos como DTOs públicos em `loam_loader` por conveniência experimental. (Resolvido em 2025-12-16 movendo para `internal/dto`).
- **2025-12-16**: *Refactoring*. Extração de `NodeMetadata` e `LoaderTransition` para `internal/dto` para limpar a API do adapter e centralizar definições.
- **2025-12-14**: *Test Strategy*. Decidido que a cobertura de testes deve ser explícita em cada fase crítica.
- **2026-01-11**: *Interpolation Strategy*. Adotada Interface `Interpolator` para permitir plugabilidade de estratégias de template (o usuário pode escolher entre Go Template, Legacy ou outros), mantendo o Core agnóstico.
- **2026-01-13**: *Tool Definition Strategy*. Adotada abordagem polimórfica para a chave `tools`. Aceita tanto definições inline (Maps) quanto referências (Strings). Decidido aceitar o trade-off de tipagem em `[]any` em troca de DX superior, mitigando riscos com validação manual e detecção de ciclos no Loader.
- **2026-01-14**: *Context Security*. Implementado namespace reservado `sys.*` no Engine. Escrita via `save_to` é bloqueada para prevenir injeção de estado. Leitura via templates é permitida para introspecção e error handling.
- **2026-01-14**: *Execution Lifecycle*. Refatorado `Engine.Navigate` para seguir estritamente `applyInput` (Update) -> `resolveTransition` (Resolve) -> `Transition`. Adicionado Deep Interpolation para argumentos de ferramenta em `Engine.Render`.
- **2026-01-15**: *Strategic Pivot*. Roadmap v0.5.2 reorientado de "Ops" para "Control & Safety". Decidido que instrumentação (Prometheus/Log) é responsabilidade do Host via Lifecycle Hooks, mantendo o Core leve. "Instrumented Adapters" removido do roadmap, com `examples/structured-logging` servindo como referência canônica.
- **2026-01-15**: *Sober Refactor*. Consolidação da confiabilidade do Runner. Unificada a lógica de nós terminais (garantindo logs de saída) e extraído `SignalManager` para isolar complexidade de concorrência. Adotado `log/slog` padronizado em todo o CLI.
- **2026-01-16**: *Roadmap Pivot*. v0.6 redefinida de "DX/Ergonomics" para "Integration & Persistence". Reconhecimento de que a gestão de estado persistente e concorrência é o "Elo Perdido" para adoção em ChatOps reais, priorizando-o sobre features de luxo (LSP/DSL).
- **2026-01-16**: *Future Phases*. Roadmap v0.7 e v0.8 reestruturados para separar preocupações de Runtime/Escala (v0.7 - Network) das preocupações de Ferramental/Ecossistema (v0.8 - Modularity).
- **2026-01-16**: *Runner Refactor Decision*. Decidido refatorar o `Runner` para usar **Functional Options Pattern**. Motivo: A injeção de `Store` e `SessionID` via argumentos/propriedades tornou a API frágil e inconsistente ("bêbada"). A configuração deve ser imutável no momento da construção.
- **2026-01-22**: *Runner Loop Simplification*. Removida otimização prematura ("Short Circuit") para nós terminais. Decisão: O `Runner` deve sempre delegar ao `Engine.Navigate` para garantir que eventos de ciclo de vida (`OnNodeLeave`) sejam disparados consistentemente, mesmo na saída.
- **2026-01-22**: *Explicit Naming Strategy*. Adotada convenção "Manual X" (`manual-saga`, `manual-security`) para exemplos que demonstram wiring explícito de features que futuramente serão nativas/automáticas. Isso preserva o espaço semântico e educa o usuário sobre a diferença entre "Padrão Nativo" e "Implementação via Código".

---
