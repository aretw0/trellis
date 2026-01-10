# Trellis

[![Go Report Card](https://goreportcard.com/badge/github.com/aretw0/trellis)](https://goreportcard.com/report/github.com/aretw0/trellis)
[![Go Doc](https://godoc.org/github.com/aretw0/trellis?status.svg)](https://godoc.org/github.com/aretw0/trellis)
[![License](https://img.shields.io/github/license/aretw0/trellis.svg)](LICENSE.txt)
[![Release](https://img.shields.io/github/release/aretw0/trellis.svg?branch=main)](https://github.com/aretw0/trellis/releases)

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é um **Motor de Máquina de Estados Determinístico** (Deterministic State Machine Engine) para a construção de CLIs, fluxos de automação e Guardrails para Agentes de IA.

Ele atua como a espinha dorsal lógica do seu sistema: enquanto sua interface (ou LLM) gerencia a conversa, o Trellis impõe estritamente as regras de negócio e as transições permitidas.

## Principais Features

- **Strict Typing & Serialization**: Garante que seus fluxos sejam robustos e livres de erros de digitação.
- **Namespaces (Sub-Grafos)**: Organize fluxos complexos em pastas e módulos (`jump_to`), escalando sua arquitetura.
- **MCP Server**: Integração nativa com **Model Context Protocol** para conectar Agentes de IA (Claude, etc.).
- **Hexagonal Architecture**: Core agnóstico (Go Library) desacoplado de FileSystem (Loam Adapter).
- **Hot Reload**: Desenvolva com feedback instantâneo.

## Quick Start

### Instalação

```bash
git clone https://github.com/aretw0/trellis
cd trellis
go mod tidy
```

### Rodando o Golden Path (Demo)

```bash
# Execução do Engine (Demo)
go run ./cmd/trellis ./examples/tour
```

## Usage

### Running a Flow

```bash
# Interactive mode
trellis run ./my-flow

# HTTP Server Mode (Stateless)
trellis serve --dir ./my-flow --port 8080
# Swagger UI available at: http://localhost:8080/swagger

# MCP Server Mode (Claude Desktop / Agent)
trellis mcp --dir ./my-flow
# Or with SSE:
trellis mcp --dir ./my-flow --transport sse --port 8080
```

### Introspection

Visualize your flow as a Mermaid graph:

```bash
trellis graph ./my-flow
# Outputs: graph TD ...
```

### Development Mode

**Using Makefile (Recommended):**

```bash
make gen    # Generate Go code from OpenAPI spec
make serve  # Run server with 'tour' example
make test   # Run tests
```

**Manual Hot Reload:**
Iterate faster on your flows by watching for file changes:

```bash
trellis run --watch --dir ./my-flow
```

The engine will monitor your `.md`, `.json`, `.yaml`, and `.yml` files. When you save a change, the session will automatically reload (preserving the workflow loop).

## Documentação

- [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
- [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
- [🌐 Guide: Running HTTP Server (Swagger)](./docs/guides/running_http_server.md)
- [🎮 Guide: Interactive Inputs](./docs/guides/interactive_inputs.md)
- [📅 Roadmap & Planning](./docs/PLANNING.md)

## Estrutura

```text
trellis/
├── cmd/           # Entrypoints (trellis CLI)
├── docs/          # Documentação do Projeto
├── examples/      # Demos e Receitas (Tours, Patterns)
├── internal/      # Implementação Privada (Loam Adapter, Runtime)
├── pkg/           # Contratos Públicos (Facade, Domain, Ports)
└── tests/         # Testes de Integração (Certification Suite)
```

## Library vs Framework

O Trellis foi desenhado para ser usado de duas formas:

1. **Como Framework (CLI)**: Use o executável `trellis` para rodar pastas de Markdown (`loam`). Ótimo para scripts rapidos e prototipagem.
2. **Como Biblioteca (Go)**: Importe `github.com/aretw0/trellis` e instancie o Engine dentro do seu binário. Você pode injetar grafos em memória, usar banco de dados ou qualquer outra fonte, sem depender de arquivos ou do Loam.

> "Opinionated by default (Loam), flexible under the hood (Memory/Custom)."
