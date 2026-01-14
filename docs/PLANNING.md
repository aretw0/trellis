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

### 🧠 v0.5: Semantic Cleanup (The "Pure" Phase)

Foco: Remover heurísticas de CLI do Core Engine e alinhar tipos de nós com semântica de State Machine pura.

- [ ] **Non-Blocking Text**: Alterar semântica padrão de `type: text` para "Pass-through" (não bloqueia).
- [ ] **Explicit Inputs**: Introduzir `type: prompt` ou `wait: true` para nós que exigem pausa/input.
- [ ] **Input Binding**: Suporte a `input_bind: "variable_name"` para salvar dados no `State.Context` sem uso de Tools.
- [ ] **Error Handling**: Adicionar transição explícita `on_error: "node_id"` para recuperação automática de falhas em Tools.
- [ ] **Observability Hooks**: Refatorar Engine para emitir eventos (`OnTransition`, `OnNodeEnter`) permitindo instrumentação externa (OpenTelemetry).
- [ ] **Migration Tool**: Utilitário para atualizar grafos antigos (`text` -> `prompt` ou adicionar `wait: true`).

### 🔮 Backlog / Concepts

- **Go DSL / Builders**: Criar helpers (`pkg/dsl` ou `pkg/builder`) para facilitar a criação de grafos em Go puro, reduzindo a verbosidade de instanciar structs manualmente e preenchendo o gap de DX em relação ao Markdown.
- **Granular SSE Events**: Evoluir de `reload` genérico para `state_changed` ou `update:<file_id>` para permitir updates parciais no frontend.
- **WASM Playground**: Compilar Trellis para WebAssembly para editor visual online.
- **Language Server Protocol (LSP)**: Plugin de VSCode para autocompletar nomes de nós e variáveis no Markdown.
- **Form Wizard Pattern**: Criar exemplo robusto de coleta de dados em múltiplos passos (Wizard) demonstrando validação e submissão, após v0.5.
- **Data Schema Validation**: Permitir definição de `required_context` no início do grafo para Fail Fast.
- **gRPC**: Interface para alta performance interna.
- **Visual Assets**: GIFs demonstrando fluxo TUI e Hot Reload no README.
- **MCP Enrichment**:
  - **Prompts Support**: Implementar `prompts/list` para iniciar fluxos comuns (ex: "Start Tour", "Debug Auth").
  - **Granular Resources**: Permitir subscrição a arquivos específicos (`trellis://file/start.md`).
  - **Sampling**: Permitir que o servidor solicite completions ao cliente (Human/LLM in the loop).
  - **Docker**: Imagem oficial para deployment fácil do servidor MCP.

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

---

## 3. Estratégia de Testes

Para evitar regressões, definimos níveis de teste obrigatórios:

1. **Core/Logic (Engine)**: Unit Tests + Table Driven.
2. **Adapters (Loam/Memory)**: *Contract Tests*. O mesmo suite deve rodar contra Loam e MemoryLoader para garantir funcionalidade idêntica.
3. **Integration**: Testes End-to-End simulando JSON in/out.
4. **CLI**: Snapshot testing.

---
