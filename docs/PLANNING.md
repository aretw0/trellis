# Planning: Trellis

> Para filosofia e arquitetura, [consulte o README](./README.md).

## Roadmap

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

### 🚧 v0.3: Developer Experience (DX) & Tooling

Foco: Ferramentas para quem *constrói* os fluxos (Toolmakers). Garantir confiança e testabilidade.

- [x] **Public Facade (Root Package)**: Refatorar para expor API no root `github.com/aretw0/trellis`.
- [x] **CLI & Runner Architecture**: Extrair loop para `Runner` e adotar `spf13/cobra` para gerenciar comandos (`run`, `graph`, `validate`).
- [x] **Compiler Validation**: O Compiler deve validar links mortos. (De-prioritized for CLI focus).
- [x] **Delegated Logic Integration**: Suporte a condicionais (`condition: is_vip`) e interpolação simples. A lógica real reside em callbacks no código Go (Host), não no Markdown.
- [ ] **Introspection (Graphviz/Mermaid)**: Comando `trellis graph` para exportar a visualização do fluxo. "Documentation as Code".
- [ ] **Headless Runner**: Capacidade de executar fluxos sem interface visual para testes automatizados de regressão.

### 🎨 v0.4: User Experience (The "Pretty" Phase)

Foco: Experiência visual do usuário final no Terminal.

- [ ] **TUI Renderer**: Integração com `charmbracelet/glamour` para renderizar Markdown rico (tabelas, alertas) no terminal.
- [ ] **Interactive Inputs**: Suporte nativo a diferentes tipos de input no frontmatter (ex: password masking, select lists, multiline text).

### � v0.5: Scale & Protocol (The "System" Phase)

Foco: Arquitetura para sistemas complexos e distribuídos.

- [ ] **Sub-Grafos (Namespaces)**: Capacidade de um nó apontar para outro arquivo/grafo (`jump_to: "checkout_flow.md"`). Permite modularização.
- [ ] **Stateless Server Mode**: Adaptador para rodar o Engine via API (HTTP/gRPC/Lambda), onde o estado é externo (Redis/Client-side).
- [ ] **Side-Effect Protocol**: Padronização de como o Trellis solicita ações ao Host (ex: retornar struct `Action` estruturada para envio de email ou DB update).

### 🔮 Backlog / Concepts

- **WASM Playground**: Compilar Trellis para WebAssembly para editor visual online.
- **Language Server Protocol (LSP)**: Plugin de VSCode para autocompletar nomes de nós e variáveis no Markdown.

---

## 3. Decisões Arquiteturais (Log)

- **2025-12-11**: *Presentation Layer Responsibility*. Decidido que a limpeza de whitespace (sanitização de output) é responsabilidade da camada de apresentação (CLI), não do Storage (Loam) ou do Domain (Engine).
- **2025-12-11**: *Loam Integration*. Adotado `TypedRepository` para mapear frontmatter automaticamente, tratando o Loam como fonte da verdade para formatos.
- **2025-12-13**: *Logic Decoupling*. Adotada estratégia de "Delegated Logic". O Markdown declara *intenções* de lógica, o Host implementa.
- **2025-12-13**: *Encapsulation*. `NodeMetadata` e `LoaderTransition` mantidos como DTOs públicos em `loam_loader` por conveniência experimental. **FIXME**: Torná-los privados ou movê-los para `internal/dto` para evitar poluição de API.
