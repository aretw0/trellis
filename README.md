# Trellis

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é um **Motor de Máquina de Estados Determinístico** (Deterministic State Machine Engine) para a construção de CLIs, fluxos de automação e Guardrails para Agentes de IA.

Ele atua como a espinha dorsal lógica do seu sistema: enquanto sua interface (ou LLM) gerencia a conversa, o Trellis impõe estritamente as regras de negócio e as transições permitidas.

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

# Headless mode (for automation/pipes)
echo "start\nyes" | trellis run --headless ./my-flow
```

### Introspection

Visualize your flow as a Mermaid graph:

```bash
trellis graph ./my-flow
# Outputs: graph TD ...
```

### Development Mode (Hot Reload)

Iterate faster on your flows by watching for file changes:

```bash
trellis run --watch --dir ./my-flow
```

The engine will monitor your `.md`, `.json`, `.yaml`, and `.yml` files. When you save a change, the session will automatically reload (preserving the workflow loop).

## Documentação

- [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
- [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
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
