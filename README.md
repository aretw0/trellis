# Trellis

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/trellis)](https://goreportcard.com/report/github.com/aretw0/trellis)
[![Go Doc](https://godoc.org/github.com/aretw0/trellis?status.svg)](https://godoc.org/github.com/aretw0/trellis)
[![License](https://img.shields.io/github/license/aretw0/trellis.svg)](LICENSE.txt)
[![Release](https://img.shields.io/github/release/aretw0/trellis.svg?branch=main)](https://github.com/aretw0/trellis/releases)

> "Simplifique o Caos. Construa fluxos determinísticos." - Filosofia Trellis

**Trellis** é um **Motor de Máquina de Estados Determinístico** (Deterministic State Machine Engine) projetado para a construção de CLIs, **ChatOps** resilientes e Guardrails para Agentes de IA (**Neuro-Symbolic**).

Ele atua como a espinha dorsal lógica do seu sistema: enquanto sua interface (ou LLM) gerencia a conversa, o Trellis impõe estritamente as regras de negócio, o contexto e as transições permitidas.

> **Visão**: O Trellis almeja ser o "Temporal Visual" — uma plataforma de **Durable Execution** que permite fluxos de longa duração (SAGA) e recuperação de falhas.

## Como funciona?

O Trellis define fluxos através de arquivos Markdown (Loam). Texto, Lógica e Dados vivem juntos:

```yaml
# start.md
type: question
text: "Olá! Qual é o seu nome?"
save_to: "user_name" # Data Binding automático
---
# greeting.md
type: text
text: "Prazer, {{ .user_name }}! O que deseja fazer?"
options: # Transições explícitas
  - text: "Ver Menu"
    to: "menu"
  - text: "Sair"
    to: "exit"
```

## Funcionalidades Principais

- **Data Binding & Contexto**: Capture inputs (`save_to`) e use variáveis (`{{ .name }}`) nativamente.
- **Namespaces (Sub-Grafos)**: Organize fluxos complexos em pastas e módulos (`jump_to`), escalando sua arquitetura.
- **MCP Server**: Integração nativa com **Model Context Protocol** para conectar Agentes de IA (Claude, Cursor, etc.).
- **Strict Typing**: Garante que seus fluxos sejam robustos e livres de erros de digitação (Zero "undefined" errors).
- **Embeddable & Agnostic**: Use como CLI, Lib ou Service. O Core é desacoplado de IO e Persistência (Hexagonal).
- **Error Handling**: Mecanismo nativo de recuperação (`on_error`) para ferramentas que falham.
- **Native SAGA Support**: Orquestração de transações distribuídas com `undo` e `rollback` automático.
- **Hot Reload**: Desenvolva com feedback instantâneo (SSE) ao salvar seus arquivos.

## Quick Start

### Instalação

Como o Trellis é tanto uma Library quanto um CLI, recomendamos clonar para ter acesso aos exemplos e ferramentas:

```bash
git clone https://github.com/aretw0/trellis
cd trellis
go mod tidy # Sincroniza dependências
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

- [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
- [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
- [🌐 Guide: Running HTTP Server (Swagger)](./docs/guides/running_http_server.md)
- [🎮 Guide: Interactive Inputs](./docs/guides/interactive_inputs.md)
- [💾 Guide: Session Management (Chaos Control)](./docs/guides/session_management.md)
- [📅 Roadmap & Planning](./docs/PLANNING.md)

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

## Library vs Framework

O Trellis foi desenhado para ser usado de duas formas:

1. **Como Framework (CLI)**: Use o executável `trellis` para rodar pastas de Markdown (`loam`). Ótimo para scripts rapidos e prototipagem.
2. **Como Biblioteca (Go)**: Importe `github.com/aretw0/trellis` e instancie o Engine dentro do seu binário. Você pode injetar grafos em memória, usar banco de dados ou qualquer outra fonte, sem depender de arquivos ou do Loam.

> "Opinionated by default (Loam), flexible under the hood (Memory/Custom)."
