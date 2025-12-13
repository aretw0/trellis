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

### 🚧 v0.3: UX & Compiler Features

Foco: Validação de grafo e experiência visual.

- [ ] **Compiler Validation**: O Compiler deve validar links mortos (`to_node_id` que não existe).
- [ ] **Renderização Rica no CLI**: Usar uma lib de TUI (ex: `charmbracelet/glamour`) para renderizar o Markdown bonito no terminal.
- [ ] **Delegated Logic Integration**: Suporte a condicionais via callbacks ("Flags de Recurso") e interpolação simples (`{{ variavel }}`). **Constraint**: Sem expressões complexas no Markdown.
- [ ] **Public Facade**: Refatorar `pkg/trellis` para expor API limpa e usar nos testes (Dogfooding), com cuidado para não complicar a importação simples e.g. `import "github.com/aretw0/trellis"`.

### 🔮 Backlog / Future

- **Sub-grafos**: Capacidade de um nó apontar para outro grafo inteiro.
- **Plugins de Ação**: Definir um padrão para ações customizadas além de `CLI_PRINT`.
- **Server Mode**: Expor o Engine via API HTTP/gRPC.

---

## 3. Decisões Arquiteturais (Log)

- **2025-12-11**: *Presentation Layer Responsibility*. Decidido que a limpeza de whitespace (sanitização de output) é responsabilidade da camada de apresentação (CLI), não do Storage (Loam) ou do Domain (Engine).
- **2025-12-11**: *Loam Integration*. Adotado `TypedRepository` para mapear frontmatter automaticamente, tratando o Loam como fonte da verdade para formatos.
