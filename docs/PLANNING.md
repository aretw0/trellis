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

### 🚧 v0.2: Hardening & UX

Foco: Segurança, validação e experiência de uso.

- [ ] **Compiler Validation**: O Compiler deve validar links mortos (`to_node_id` que não existe).
- [ ] **Renderização Rica no CLI**: Usar uma lib de TUI (ex: `charmbracelet/glamour`) para renderizar o Markdown bonito no terminal.
- [ ] **Variáveis e Lógica**: Suporte a interpolação simples (ex: `Olá {{ nome }}`) e condicionais mais ricas.
- [ ] **Testes de Unidade**: Cobrir o Compiler e casos de borda do Engine.

### 🔮 Backlog / Future

- **Sub-grafos**: Capacidade de um nó apontar para outro grafo inteiro.
- **Plugins de Ação**: Definir um padrão para ações customizadas além de `CLI_PRINT`.
- **Server Mode**: Expor o Engine via API HTTP/gRPC.

---

## 3. Decisões Arquiteturais (Log)

- **2025-12-11**: *Presentation Layer Responsibility*. Decidido que a limpeza de whitespace (sanitização de output) é responsabilidade da camada de apresentação (CLI), não do Storage (Loam) ou do Domain (Engine).
- **2025-12-11**: *Loam Integration*. Adotado `TypedRepository` para mapear frontmatter automaticamente, tratando o Loam como fonte da verdade para formatos.
