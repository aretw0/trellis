# Trellis

> "Faça uma coisa e faça bem feita. Trabalhe com fluxos de texto." - Filosofia Unix

**Trellis** é o "Cérebro Lógico" de um sistema de automação. Projetada como uma **Função Pura de Transição de Estado**, opera isolada de efeitos colaterais.

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

## Documentação

- [📖 Product Vision & Philosophy](./docs/PRODUCT.md)
- [🏗 Architecture & Technical Details](./docs/TECHNICAL.md)
- [📅 Roadmap & Planning](./docs/PLANNING.md)

## Estrutura

```text
trellis/
├── cmd/           # Entrypoints (trellis, gen-trail)
├── docs/          # Documentação do Projeto
├── internal/      # Implementação (Loam Adapter, Runtime)
└── pkg/           # Contratos Públicos (Domain, Ports)
```
