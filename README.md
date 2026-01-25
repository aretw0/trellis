# Trellis

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/trellis)](https://goreportcard.com/report/github.com/aretw0/trellis)
[![Go Doc](https://godoc.org/github.com/aretw0/trellis?status.svg)](https://godoc.org/github.com/aretw0/trellis)
[![License](https://img.shields.io/github/license/aretw0/trellis.svg)](LICENSE.txt)
[![Release](https://img.shields.io/github/release/aretw0/trellis.svg?branch=main)](https://github.com/aretw0/trellis/releases)

> **The Neuro-Symbolic Backbone for Agents & Automation.**

**Trellis** é um **Motor de Máquina de Estados Determinístico** (Deterministic State Machine Engine) para a construção de CLIs, **ChatOps** resilientes e Guardrails para Agentes de IA (**Neuro-Symbolic**).

Atuando como a espinha dorsal lógica do sistema, ele impõe estritamente as regras de negócio e transições permitidas, enquanto sua interface (ou LLM) gerencia apenas a apresentação.

Mais do que um engine, é uma plataforma de **Durable Execution** que permite a suspensão e retomada de processos longos, habilitando padrões avançados como **SAGA** (Orquestração de Transações e Compensação).

> **Hybrid Nature**: Use como **Framework** (CLI + Markdown) para prototipagem rápida, ou como **Biblioteca** (Go) para controle total em seu backend. *"Opinionated by default, flexible under the hood."*

## O Conceito Neuro-Simbólico & Automação

O Trellis preenche a lacuna entre a **Rigidez dos Processos** e a **Flexibilidade da Inteligência**:

* **Para Agentes de IA**: Substitua "If/Else" frágeis e Prompts gigantes por um grafo de estados auditável. O Trellis impede alucinações de fluxo.
* **Para Humanos**: Funciona como um motor de **Workflow as Code** (similar a um n8n/Zapier, mas compilado e versionável), ideal para CLIs complexas e automação de Ops.

```mermaid
graph TD
    %% Nodes
    Brain["🧠 Cérebro (LLM) ou<br/>👤 Humano (CLI)"] -->|Intenção / Input| Trellis["🛡️ Espinha Dorsal<br/>(Trellis Engine)"]
    
    subgraph "Mundo Simbólico (Determinístico)"
        Trellis -->|Validação| Rules["📜 Regras de Negócio<br/>(State Machine)"]
        Rules -->|Ok / Block| Trellis
    end
    
    Trellis -->|Execução Segura| Tools["⚡ Ferramentas<br/>(API / DB / Scripts)"]
    Tools -->|Resultado| Trellis
    Trellis -->|Contexto Atualizado| Brain

    %% Styles
    style Brain fill:#f9f,stroke:#333,stroke-width:2px,color:black
    style Trellis fill:#9cf,stroke:#333,stroke-width:2px,color:black
    style Rules fill:#ff9,stroke:#333,stroke-width:2px,stroke-dasharray: 5 5,color:black
    style Tools fill:#9f9,stroke:#333,stroke-width:2px,color:black
```

O decisor (seja **IA** ou **Humano**) escolhe *qual* caminho tomar, mas o Trellis garante que ele *existe* e é *válido*.

## Como funciona?

Coreografamos sua lógica em um **Grafo de Estados**. Você define **Nós** (Passos) e **Transições** (Regras), e o Trellis gerencia a navegação.

Você pode definir esse grafo de duas formas:

### 1. Declarativo (Arquivos)

Ideal para prototipagem, visualização (Mermaid) e edição por LLMs. Suporta **Markdown** (Frontmatter), **YAML** ou **JSON** via [Loam](https://github.com/aretw0/loam).

```yaml
# start.yaml
type: question
content: Olá! Qual é o seu nome?
save_to: user_name  # Data Binding automático
to: greeting        # Transição incondicional
```

### 2. Programático (Go Structs)

Ideal para integração profunda em backends, performance crítica e type-safety total.

```go
&domain.Node{
    ID: "start",
    Type: "question",
    Content: []byte("Olá! Qual é o seu nome?"),
    SaveTo: "user_name",
    Transitions: []domain.Transition{{ToNodeID: "greeting"}},
}
```

> **Nota**: Ambas as formas geram a mesma estrutura em memória e podem co-existir (ex: carregar arquivos e injetar nós via código).

## Funcionalidades Principais

* **Data Binding & Contexto**: Capture inputs (`save_to`) e use variáveis (`{{ .name }}`) nativamente.
* **Namespaces (Sub-Grafos)**: Organize fluxos complexos em pastas e módulos (`jump_to`), escalando sua arquitetura.
* **MCP Server**: Integração nativa com **Model Context Protocol** para conectar Agentes de IA (Claude, Cursor, etc.).
* **Strict Typing**: Garante que seus fluxos sejam robustos e livres de erros de digitação (Zero "undefined" errors).
* **Embeddable & Agnostic**: Use como CLI, Lib ou Service. O Core é desacoplado de IO e Persistência (Hexagonal).
* **Error Handling**: Mecanismo nativo de recuperação (`on_error`) para ferramentas que falham.
* **Native SAGA Support**: Orquestração de transações distribuídas com `undo` e `rollback` automático.
* **Hot Reload**: Desenvolva com feedback instantâneo (SSE) ao salvar seus arquivos.

## Quick Start

### Instalação

#### Windows (Recomendado)

A forma mais fácil de instalar no Windows é via **Scoop**:

```powershell
# 1. Adicione o bucket (apenas a primeira vez)
scoop bucket add aretw0 https://github.com/aretw0/scoop-bucket

# 2. Instale o Trellis
scoop install trellis
```

#### macOS / Linux

Instale via **Homebrew**:

```bash
brew install aretw0/tap/trellis
```

#### Via Go (Library Mode)

Para usar o Trellis como biblioteca dentro do seu backend (sem arquivos, puramente em memória):

```bash
go get github.com/aretw0/trellis
```

```go
// Exemplo: Instanciando o Engine sem ler arquivos
loader, _ := memory.NewFromNodes(myNodes...)
eng, _ := trellis.New("", trellis.WithLoader(loader))
```

### Rodando o Golden Path (Demo)

```bash
# Execução do Engine (Demo)
go run ./cmd/trellis ./examples/tour
```

## Usage

### Rodando um Fluxo (CLI)

```bash
# Modo Interativo (Terminal)
go run ./cmd/trellis run ./examples/tour

# Modo HTTP Server (Stateless API)
go run ./cmd/trellis serve --dir ./examples/tour --port 8080
# Swagger UI disponível em: http://localhost:8080/swagger

# Modo MCP Server (Para Agentes de IA)
go run ./cmd/trellis mcp --dir ./examples/tour

# Modo Debug (Observability)
go run ./cmd/trellis run --debug ./examples/observability

# Exemplo Global Signals (Interrupts)
go run ./cmd/trellis run ./examples/interrupt-demo

# Exemplo Tool Safety & Error Handling
go run ./cmd/trellis run ./examples/tools-demo

# Exemplo Log Estruturado (Production Recipe)
go run ./examples/structured-logging
```

### Introspecção

Visualize seu fluxo como um grafo Mermaid:

```bash
trellis graph ./my-flow
# Saída: graph TD ...
```

### Modo de Desenvolvimento

**Usando Makefile (Recomendado):**

```bash
make gen    # Gera código Go a partir da spec OpenAPI
make serve  # Roda servidor com exemplo 'tour'
make test   # Roda testes
```

**Hot Reload Manual:**
Itere mais rápido observando mudanças de arquivo:

```bash
trellis run --watch --dir ./my-flow
```

O engine monitorará seus arquivos `.md`, `.json`, `.yaml`. Ao salvar, a sessão recarrega automaticamente (preservando o loop de execução).

## Documentação

* [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
* [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
* [🌐 Guide: Running HTTP Server (Swagger)](./docs/guides/running_http_server.md)
* [🎮 Guide: Interactive Inputs](./docs/guides/interactive_inputs.md)
* [💾 Guide: Session Management (Chaos Control)](./docs/guides/session_management.md)
* [📅 Roadmap & Planning](./docs/PLANNING.md)
* [⚖️ Decisões de Arquitetura](./docs/DECISIONS.md)

## Estrutura

```text
trellis/
├── cmd/           # Entrypoints (trellis CLI)
├── docs/          # Documentação do Projeto
├── examples/      # Demos e Receitas (Tours, Patterns)
├── internal/      # Implementação Privada (Runtime, TUI)
├── pkg/           # Contratos Públicos (Facade, Domain, Ports, Adapters)
└── tests/         # Testes de Integração (Certification Suite)
```

## Licença

[AGPL-3.0](LICENSE.txt)
